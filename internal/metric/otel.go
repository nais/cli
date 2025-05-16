package metric

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nais/cli/internal/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	CliName      = "nais_cli"
	collectorURL = "https://collector-internet.nav.cloud.nais.io"
)

var initialized = false

func Initialize() func(verbose bool) {
	if os.Getenv("DO_NOT_TRACK") == "1" || initialized {
		return func(verbose bool) {
			if verbose {
				fmt.Printf("Shutdown: skipping metrics upload as DO_NOT_TRACK is 1..\n")
			}
		}
	}

	initialized = true

	provider := newMeterProvider()
	otel.SetMeterProvider(provider)

	return func(verbose bool) {
		if verbose {
			fmt.Printf("Shutdown: uploading metrics...\n")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := provider.Shutdown(ctx)
		if err != nil {
			fmt.Printf("Failed up upload metrics: %v\n", err)
		}
	}
}

func CreateCounter(metricName string, attributes ...attribute.KeyValue) m.Int64Counter {
	meter := otel.GetMeterProvider().Meter(CliName)
	counter, _ := meter.Int64Counter(CliName+"_"+metricName, m.WithUnit("1"))

	return counter
}

func CreateAndIncreaseCounter(ctx context.Context, metricName string, attributes ...attribute.KeyValue) {
	counter := CreateCounter(metricName, attributes...)
	counter.Add(ctx, 1)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(CliName),
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
