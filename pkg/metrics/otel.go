package metrics

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/sdk/resource"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	version           = "local"
	commit            = "uncommited"
	naisCliPrefixName = "nais_cli"
	collectorURL      = "https://collector-internet.nav.cloud.nais.io"
)

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("nais_cli"),
			semconv.ServiceVersion(version+"-"+commit),
		))
}

func newMeterProvider(res *resource.Resource) *metric.MeterProvider {
	dnt := os.Getenv("DO_NOT_TRACK")
	url := collectorURL
	if dnt == "1" {
		url = "http://localhost:1234"
	}
	metricExporter, _ := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpointURL(url),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(1*time.Second))),
	)
	return meterProvider
}

func recordCommandUsage(ctx context.Context, provider *metric.MeterProvider, flags []string) {
	commandHistogram, _ := provider.Meter(naisCliPrefixName).Int64Histogram(
		naisCliPrefixName+"_command_usage",
		m.WithUnit("1"),
		m.WithDescription("Usage frequency of command flags"))
	if flags != nil {
		commandHistogram.Record(ctx, 1, m.WithAttributes(attribute.String("command", flags[0])))
	}
	for i, f := range flags {
		if i == 0 {
			continue
		}
		commandHistogram.Record(ctx, 1, m.WithAttributes(attribute.String("subcommand", f)))
	}
}

// intersection
// Just a list intersection, used to create the intersection
// between os.args and all the args we have in the cli
func intersection(list1, list2 []string) []string {
	elements := make(map[string]bool)
	resultSet := make(map[string]bool)

	// Mark elements in list1
	for _, item := range list1 {
		elements[item] = true
	}

	// Check for intersections and add to resultSet to ensure uniqueness
	for _, item := range list2 {
		if elements[item] && !resultSet[item] {
			resultSet[item] = true
		}
	}

	// Collect the unique intersection elements into a slice
	var result []string
	for item := range resultSet {
		result = append(result, item)
	}

	return result
}

func CollectCommandHistogram(commands []*cli.Command) {
	doNotTrack := os.Getenv("DO_NOT_TRACK")
	if doNotTrack == "1" {
		log.Default().Println("DO_NOT_TRACK is set, not collecting metrics")
	}

	ctx := context.Background()
	var validSubcommands []string
	for _, command := range commands {
		gatherCommands(command, &validSubcommands)
	}

	res, _ := newResource()
	provider := newMeterProvider(res)

	// Record usages of subcommands that are exactly in the list of args we have, nothing else
	recordCommandUsage(ctx, provider, intersection(os.Args, validSubcommands))
	provider.Shutdown(ctx)
}

func gatherCommands(command *cli.Command, validSubcommands *[]string) {
	*validSubcommands = append(*validSubcommands, command.Name)
	for _, subcommand := range command.Subcommands {
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
func AddOne(metricName string) {
	ctx := context.Background()
	counterName := naisCliPrefixName + "_" + metricName
	res, _ := newResource()
	meter := newMeterProvider(res)
	counter, _ := meter.Meter(naisCliPrefixName).Int64Counter(
		counterName,
		m.WithUnit("1"),
		m.WithDescription("Counter for "+counterName),
	)

	counter.Add(ctx, 1, m.WithAttributes(attribute.String("command", metricName)))
	defer meter.Shutdown(ctx)
}
