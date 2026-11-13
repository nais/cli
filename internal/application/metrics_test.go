package application

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/nais/naistrix"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestCommandTelemetryCoverage(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		command    string
		success    bool
		completion bool
	}{
		{name: "unknown command", args: []string{"secret-user-input"}, command: "unknown"},
		{name: "unknown nested command", args: []string{"app", "secret-user-input"}, command: "unknown"},
		{name: "help", args: []string{"--help"}, command: "nais", success: true},
		{name: "help command", args: []string{"help", "app"}, command: "nais help", success: true},
		{name: "version", args: []string{"--version"}, command: "nais", success: true},
		{name: "nested help", args: []string{"app", "list", "--help"}, command: "nais app list", success: true},
		{name: "failed real command", args: []string{"app", "list", "--invalid-flag"}, command: "nais app list"},
		{name: "completion", args: []string{"__complete", ""}, completion: true},
		{name: "completion without descriptions", args: []string{"__completeNoDesc", ""}, completion: true},
		{name: "completion script", args: []string{"completion", "bash"}, completion: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, _, err := newApplication(io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--config", filepath.Join(t.TempDir(), "config.yaml")}, tc.args...)
			runErr := app.Run(naistrix.RunWithArgs(args))
			cmds := app.ExecutedCommand()
			if got := isCompletionCommand(cmds); got != tc.completion {
				t.Fatalf("completion = %v, want %v; command = %v; error = %v", got, tc.completion, cmds, runErr)
			}
			if tc.completion {
				// Completion must be excluded even when configuration or auth fails.
				return
			}

			reader := metric.NewManualReader()
			provider := metric.NewMeterProvider(metric.WithReader(reader))
			previous := otel.GetMeterProvider()
			otel.SetMeterProvider(provider)
			t.Cleanup(func() {
				otel.SetMeterProvider(previous)
				_ = provider.Shutdown(context.Background())
			})
			collectCommandMetrics(t.Context(), cmds, runErr, 1500*time.Millisecond)
			var data metricdata.ResourceMetrics
			if err := reader.Collect(t.Context(), &data); err != nil {
				t.Fatal(err)
			}
			if len(data.ScopeMetrics) != 1 || len(data.ScopeMetrics[0].Metrics) != 2 {
				t.Fatalf("expected usage and duration metrics: %+v", data)
			}
			checkAttributes := func(name string, attrs attribute.Set) {
				t.Helper()
				command, _ := attrs.Value(attribute.Key("command"))
				success, _ := attrs.Value(attribute.Key("success"))
				if command.AsString() != tc.command || success.AsBool() != tc.success {
					t.Fatalf("%s: command=%q success=%v; want command=%q success=%v; run error=%v", name, command.AsString(), success.AsBool(), tc.command, tc.success, runErr)
				}
			}
			for _, m := range data.ScopeMetrics[0].Metrics {
				switch h := m.Data.(type) {
				case metricdata.Histogram[int64]:
					if m.Name != "nais_cli_command_usage" || len(h.DataPoints) != 1 || h.DataPoints[0].Sum != 1 {
						t.Fatalf("unexpected usage metric: %+v", m)
					}
					checkAttributes(m.Name, h.DataPoints[0].Attributes)
				case metricdata.Histogram[float64]:
					if m.Name != "nais_cli_command_duration" || m.Unit != "s" || len(h.DataPoints) != 1 || h.DataPoints[0].Sum != 1.5 {
						t.Fatalf("unexpected duration metric: %+v", m)
					}
					checkAttributes(m.Name, h.DataPoints[0].Attributes)
				default:
					t.Fatalf("unexpected metric %s: %T", m.Name, m.Data)
				}
			}
		})
	}
}
