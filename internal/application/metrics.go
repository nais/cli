package application

import (
	"context"
	"strings"
	"time"

	metrics "github.com/nais/cli/internal/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func isCompletionCommand(cmds []string) bool {
	if len(cmds) < 2 {
		return false
	}
	switch cmds[1] {
	case "completion", "__complete", "__completeNoDesc":
		return true
	default:
		return false
	}
}

func collectCommandMetrics(ctx context.Context, cmds []string, err error, duration time.Duration) {
	meter := otel.GetMeterProvider().Meter(metrics.CliName)

	commandHistogram, _ := meter.Int64Histogram(
		metrics.CliName+"_command_usage",
		metric.WithUnit("1"),
		metric.WithDescription("Usage frequency of Nais CLI commands"),
	)
	durationHistogram, _ := meter.Float64Histogram(
		metrics.CliName+"_command_duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of Nais CLI commands, excluding telemetry upload"),
		metric.WithExplicitBucketBoundaries(0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 300),
	)

	command := strings.Join(cmds, " ")
	// Cobra returns only known command names, but may return a known parent
	// alongside an unknown-command error. Never include the rejected input.
	if len(cmds) == 0 || (err != nil && strings.HasPrefix(err.Error(), "unknown command ")) {
		command = "unknown"
	}
	attributes := []attribute.KeyValue{
		attribute.String("command", command),
		attribute.Bool("success", err == nil),
	}
	commandHistogram.Record(ctx, 1, metric.WithAttributes(attributes...))
	durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attributes...))
}
