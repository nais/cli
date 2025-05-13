package metrics

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	cliName      = "nais_cli"
	collectorURL = "https://collector-internet.nav.cloud.nais.io"
)

var provider *metric.MeterProvider

func init() {
	provider = newMeterProvider()
}

func CollectCommandHistogram(ctx context.Context, cmd *cobra.Command, err error) {
	if os.Getenv("DO_NOT_TRACK") == "1" {
		return
	}

	commandHistogram, _ := provider.Meter(cliName).Int64Histogram(
		cliName+"_command_usage",
		m.WithUnit("1"),
		m.WithDescription("Usage frequency of Nais CLI commands"),
	)

	attributes := []attribute.KeyValue{
		attribute.String("command", strings.Join(commandNames(cmd), " ")),
		attribute.Bool("success", err == nil),
	}
	commandHistogram.Record(ctx, 1, m.WithAttributes(attributes...))

	_ = provider.Shutdown(ctx)
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
	counterName := cliName + "_" + metricName
	counter, _ := provider.Meter(cliName).Int64Counter(
		counterName,
		m.WithUnit("1"),
		m.WithDescription("Counter for "+counterName),
	)

	counter.Add(ctx, 1, m.WithAttributes(attribute.String("command", metricName)))
	defer func() {
		_ = provider.Shutdown(ctx)
	}()
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(cliName),
			semconv.ServiceVersion(version.Version),
		),
	)
}

func newMeterProvider() *metric.MeterProvider {
	res, _ := newResource()

	metricExporter, _ := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpointURL(collectorURL),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(
			metricExporter,
			metric.WithInterval(1*time.Second)),
		),
	)
	return meterProvider
}

func commandNames(cmd *cobra.Command) []string {
	if cmd == nil {
		return nil
	}

	return append(commandNames(cmd.Parent()), cmd.Name())
}
