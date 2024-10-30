package metrics

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/sdk/resource"
	"os"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	version = "local"
	commit  = "uncommited"
)

func newMeterProvider(res *resource.Resource) *metric.MeterProvider {
	dnt := os.Getenv("DO_NOT_TRACK")
	var url string
	if dnt == "1" {
		fmt.Println("We are respecting your do-not-track")
		url = "http://localhost:1234"
	} else {
		url = "https://collector-internet.nav.cloud.nais.io"
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

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("nais_cli"),
			semconv.ServiceVersion(version+":"+commit),
		))
}

func New() *metric.MeterProvider {
	res, _ := newResource()
	meterProvider := newMeterProvider(res)
	return meterProvider
}

func RecordCommandUsage(ctx context.Context, histogram m.Int64Histogram, flags []string) {
	for _, f := range flags {
		histogram.Record(ctx, 1, m.WithAttributes(attribute.String("flag", f)))
	}

}

// Intersection
// Just a list intersection, used to create the intersection
// between os.args and all the args we have in the cli
func Intersection(list1, list2 []string) []string {
	elements := make(map[string]bool)
	for _, item := range list1 {
		elements[item] = true
	}
	var result []string
	for _, item := range list2 {
		if elements[item] {
			result = append(result, item)
		}
	}
	return result
}

// AddOne
// This calls New(), creating a whole new MeterProvider on every invocation.
// This will result in many 1s being sent as their own unique snowflake 1.
// This is because the otel.setMeterprovider/otel.getMeterProvider doesn't expose
// ForceFlush meaning we have to wait a second or so after every command to send the
// buffered metrics up. This is extremely silly and wtf. Instead we always create a new metricprovider,
// add a metric anf forceflush it. v0v
// We tried using the global set/get meterprovider but that does not give forceflush and instead
// you end up doing a sleep(2s) to get the metrics sent which is maybe not the best ux I can imagine.
func AddOne(meterName, counterName string) {
	ctx := context.Background()
	meter := New()
	counter, _ := meter.Meter(meterName).Int64Counter(counterName)
	counter.Add(ctx, 1)
	_ = meter.ForceFlush(ctx)
	defer meter.Shutdown(ctx)
}
