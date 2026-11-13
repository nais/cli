package application

import (
	"context"
	"strings"

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

func collectCommandHistogram(ctx context.Context, cmds []string, err error) {
	meter := otel.GetMeterProvider().Meter(metrics.CliName)

	commandHistogram, _ := meter.Int64Histogram(
		metrics.CliName+"_command_usage",
		metric.WithUnit("1"),
		metric.WithDescription("Usage frequency of Nais CLI commands"),
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
}
