package application

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/naistrix"
)

func TestCloudResourceCreateEnvironment(t *testing.T) {
	for _, resource := range []string{"valkey", "opensearch"} {
		for _, tt := range []struct {
			name, environment, configured string
			rejected                      bool
		}{
			{name: "dev FSS", environment: "dev-fss", rejected: true},
			{name: "prod FSS", environment: "prod-fss", rejected: true},
			{name: "FSS default", configured: "dev-fss", rejected: true},
			{name: "cloud environment", environment: "dev-gcp"},
			{name: "cloud default", configured: "prod-gcp"},
			{name: "explicit cloud overrides FSS default", environment: "dev-gcp", configured: "dev-fss"},
			{name: "explicit FSS overrides cloud default", environment: "prod-fss", configured: "prod-gcp", rejected: true},
			{name: "non NAV name", environment: "staging"},
		} {
			t.Run(resource+"/"+tt.name, func(t *testing.T) {
				t.Setenv("NAIS_ENVIRONMENT", "")
				config := filepath.Join(t.TempDir(), "config.yaml")
				contents := "{}\n"
				if tt.configured != "" {
					contents = "environment: " + tt.configured + "\n"
				}
				if err := os.WriteFile(config, []byte(contents), 0o600); err != nil {
					t.Fatal(err)
				}
				app, _, err := New(io.Discard)
				if err != nil {
					t.Fatal(err)
				}
				ran := false
				for _, cmd := range app.Commands {
					if cmd.Name != resource {
						continue
					}
					for _, sub := range cmd.SubCommands {
						if sub.Name == "create" {
							sub.RunFunc = func(context.Context, *naistrix.Arguments, *naistrix.OutputWriter) error {
								ran = true
								return nil
							}
						}
					}
				}
				args := []string{"--config", config, resource, "create", "instance", "-t", "my-team"}
				if tt.environment != "" {
					args = append(args, "-e", tt.environment)
				}
				err = app.Run(naistrix.RunWithArgs(args))
				if tt.rejected {
					if err == nil || !strings.Contains(err.Error(), "is not supported in environment") {
						t.Errorf("error = %v, want unsupported environment", err)
					}
					if ran {
						t.Error("creation reached execution in an unsupported environment")
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if !ran {
						t.Error("supported environment did not reach execution")
					}
				}
			})
		}
	}
}
