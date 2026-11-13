package metric

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nais/cli/internal/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	CliName       = "nais_cli"
	collectorURL  = "https://collector-internet.nav.cloud.nais.io/v1/metrics"
	uploadTimeout = 2 * time.Second
)

// Initialize collects locally until shutdown. Only trace mode waits for network I/O.
func Initialize() func(trace bool) {
	if os.Getenv("DO_NOT_TRACK") == "1" {
		return func(trace bool) {
			if trace {
				fmt.Fprintln(os.Stderr, "Shutdown: skipping metrics upload as DO_NOT_TRACK is 1.")
			}
		}
	}

	reader := metric.NewManualReader()
	res, _ := newResource()
	provider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	otel.SetMeterProvider(provider)

	return func(trace bool) {
		ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
		defer cancel()
		defer func() { _ = provider.Shutdown(ctx) }()

		if trace {
			fmt.Fprintln(os.Stderr, "Shutdown: uploading metrics...")
		}
		if err := export(ctx, reader, trace); err != nil && trace {
			fmt.Fprintf(os.Stderr, "Failed to upload metrics: %v\n", err)
		}
	}
}

func export(ctx context.Context, reader *metric.ManualReader, trace bool) error {
	var data metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &data); err != nil {
		return err
	}
	if len(data.ScopeMetrics) == 0 {
		return nil
	}

	client := &http.Client{Timeout: uploadTimeout}
	if !trace {
		client.Transport = backgroundTransport{}
	}
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(collectorURL),
		otlpmetrichttp.WithCompression(otlpmetrichttp.NoCompression),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: false}),
		otlpmetrichttp.WithHTTPClient(client),
	)
	if err != nil {
		return err
	}
	defer func() { _ = exporter.Shutdown(ctx) }()
	return exporter.Export(ctx, &data)
}

func CreateCounter(metricName string) m.Int64Counter {
	meter := otel.GetMeterProvider().Meter(CliName)
	counter, _ := meter.Int64Counter(CliName+"_"+metricName, m.WithUnit("1"))

	return counter
}

func CreateAndIncreaseCounter(ctx context.Context, metricName string) {
	counter := CreateCounter(metricName)
	counter.Add(ctx, 1)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(CliName),
			semconv.ServiceVersion(version.Version),
		),
	)
}
