package metrics

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	version           = "local"
	naisCliPrefixName = "nais_cli"
	collectorURL      = "https://collector-internet.nav.cloud.nais.io"
)

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("nais_cli"),
			semconv.ServiceVersion(version),
		))
}

func newMeterProvider(ctx context.Context, res *resource.Resource) *metric.MeterProvider {
	dnt := os.Getenv("DO_NOT_TRACK")
	url := collectorURL
	if dnt == "1" {
		url = "http://localhost:1234"
	}
	metricExporter, _ := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpointURL(url),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(1*time.Second))),
	)
	return meterProvider
}

func recordCommandUsage(ctx context.Context, provider *metric.MeterProvider, allCommands []string, mainCommands []*cli.Command) {
	commandHistogram, _ := provider.Meter(naisCliPrefixName).Int64Histogram(
		naisCliPrefixName+"_command_usage",
		m.WithUnit("1"),
		m.WithDescription("Usage frequency of command flags"))

	validCommands := map[string]bool{}
	for _, command := range mainCommands {
		validCommands[command.Name] = true
	}

	attributes := make([]attribute.KeyValue, 0)
	if len(allCommands) > 0 && validCommands[allCommands[0]] {
		attributes = append(attributes, attribute.String("command", allCommands[0]))
		if len(allCommands) > 1 {
			attributes = append(attributes, attribute.String("subcommand", strings.Join(allCommands[1:], "_")))
		}
	}
	commandHistogram.Record(ctx, 1, m.WithAttributes(attributes...))
}

// intersection
// Just a list intersection, used to create the intersection
// between os.args and all the args we have in the cli
func intersection(list1, list2 []string) []string {
	elements := make(map[string]bool)
	seen := make(map[string]bool)

	// Mark elements in list1
	for _, item := range list1 {
		elements[item] = true
	}

	// Check for intersections and add to resultSet to ensure uniqueness
	var result []string
	for _, item := range list2 {
		if elements[item] && !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}
	return result
}

func CollectCommandHistogram(ctx context.Context, commands []*cli.Command) {
	doNotTrack := os.Getenv("DO_NOT_TRACK")
	if doNotTrack == "1" {
		fmt.Println("DO_NOT_TRACK is set, not collecting metrics")
	}

	var validSubcommands []string
	for _, command := range commands {
		gatherCommands(command, &validSubcommands)
	}

	res, _ := newResource()
	provider := newMeterProvider(ctx, res)

	// Record usages of subcommands that are exactly in the list of args we have, nothing else
	recordCommandUsage(ctx, provider, intersection(os.Args, validSubcommands), commands)
	_ = provider.Shutdown(ctx)
}

func gatherCommands(command *cli.Command, validSubcommands *[]string) {
	*validSubcommands = append(*validSubcommands, command.Name)
	for _, subcommand := range command.Commands {
		gatherCommands(subcommand, validSubcommands) // Recursively handle subcommands
	}
}

// AddOne
// This calls NewMeterProvider(), creating a whole new MeterProvider on every invocation.
// This will result in many 1s being sent as their own unique snowflake 1.
// This is because the otel.setMeterprovider/otel.getMeterProvider doesn't expose
// ForceFlush meaning we have to wait a second or so after every command to send the
// buffered metrics up. This is extremely silly and wtf. Instead we always create a new metricprovider,
// add a metric anf forceflush it. v0v
// We tried using the global set/get meterprovider but that does not give forceflush and instead
// you end up doing a sleep(2s) to get the metrics sent which is maybe not the best ux I can imagine.
func AddOne(ctx context.Context, metricName string) {
	counterName := naisCliPrefixName + "_" + metricName
	res, _ := newResource()
	meter := newMeterProvider(ctx, res)
	counter, _ := meter.Meter(naisCliPrefixName).Int64Counter(
		counterName,
		m.WithUnit("1"),
		m.WithDescription("Counter for "+counterName),
	)

	counter.Add(ctx, 1, m.WithAttributes(attribute.String("command", metricName)))
	defer func() {
		_ = meter.Shutdown(ctx)
	}()
}
