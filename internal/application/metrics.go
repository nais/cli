package application

import (
	"context"
	"strings"

	metrics "github.com/nais/cli/internal/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func collectCommandHistogram(ctx context.Context, cmds []string, err error) {
	meter := otel.GetMeterProvider().Meter(metrics.CliName)

	commandHistogram, _ := meter.Int64Histogram(
		metrics.CliName+"_command_usage",
		metric.WithUnit("1"),
		metric.WithDescription("Usage frequency of Nais CLI commands"),
	)

	attributes := []attribute.KeyValue{
		attribute.String("command", strings.Join(cmds, " ")),
		attribute.Bool("success", err == nil),
	}
	commandHistogram.Record(ctx, 1, metric.WithAttributes(attributes...))
}
