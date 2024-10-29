package metrics

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("nais_cli"),
			// TODO: How is this number passed down?
			semconv.ServiceVersion("1.2.0"),
		))
}

func NewMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithEndpointURL("https://collector-internet.nav.cloud.nais.io"))
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(1*time.Second))),
	)
	return meterProvider, nil
}

func New() *metric.MeterProvider {
	res, err := NewResource()
	if err != nil {
		panic(err)
	}

	meterProvider, err := NewMeterProvider(res)
	if err != nil {
		panic(err)
	}

	return meterProvider

}

// This calls New(), creating a whole new MeterProvider on every invocation.
// This will result in many 1s being sent as their own unique snowflake 1.
// This is because the otel.setMeterprovider/otel.getMeterProvider doesn't expose
// ForceFlush meaning we have to wait a second or so after every command to send the
// buffered metrics up. This is extremely silly and wtf. Instead we always create a new metricprovider,
// add a metric anf forceflush it. v0v
// We tried using the global set/get meterprovider but that does not give forceflush and instead
// you end up doing a sleep(2s) to get the metrics sent which is maybe not the best ux I can imagine.
func AddOne(meterName, counterName string) error {
	myMetric := New()
	counter, err := myMetric.Meter(meterName).Int64Counter(counterName)
	counter.Add(context.Background(), 1)
	if err != nil {
		return fmt.Errorf("metrics provider")
	}
	myMetric.ForceFlush(
		context.Background())
	return nil
}
