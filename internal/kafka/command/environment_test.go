package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func newKafkaTestApplication(t *testing.T, writer io.Writer) *naistrix.Application {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("NAIS_TEAM", "")
	t.Setenv("NAIS_ENVIRONMENT", "")
	app, global, err := naistrix.NewApplication("test", "Test Kafka", "test", naistrix.ApplicationWithWriter(writer))
	if err != nil {
		t.Fatal(err)
	}
	global.Config = filepath.Join(t.TempDir(), "config.yaml")
	additional := &flags.AdditionalFlags{}
	if err := app.AddGlobalFlags(additional); err != nil {
		t.Fatal(err)
	}
	if err := app.AddCommand(Kafka(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
		t.Fatal(err)
	}
	return app
}

// TODO: Remove legacy alias rejection coverage on 2027-02-11, after the 90-day transition.
func TestKafkaTopicEnvironment(t *testing.T) {
	for _, tt := range []struct {
		name, environment, environments, wantError string
		command                                    []string
		apiError                                   bool
	}{
		{name: "list rejects alias", command: []string{"list"}, environment: "dev", wantError: `unknown Nais environment "dev"; available environments: dev-gcp, prod-gcp`},
		{name: "grant rejects alias", command: []string{"grant-access", "alice", "orders"}, environment: "dev", wantError: `unknown Nais environment "dev"`},
		{name: "list-grants rejects alias", command: []string{"list-grants", "orders"}, environment: "dev", wantError: `unknown Nais environment "dev"`},
		{name: "revoke rejects alias", command: []string{"revoke-grant", "alice", "orders", "read"}, environment: "dev", wantError: `unknown Nais environment "dev"`},
		{name: "canonical environment", environment: "dev-gcp"},
		{name: "dev is valid in another tenant", environment: "dev", environments: `[{"name":"dev"}]`},
		{name: "list without environment"},
		{name: "API error preserved", environment: "dev-gcp", apiError: true, wantError: "access denied"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct{ OperationName string }
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				switch request.OperationName {
				case "Environments":
					if tt.environment == "" {
						t.Error("listing all environments should not need environment validation")
					}
					if tt.apiError {
						_, _ = io.WriteString(w, `{"errors":[{"message":"access denied"}]}`)
						return
					}
					environments := tt.environments
					if environments == "" {
						environments = `[{"name":"prod-gcp"},{"name":"dev-gcp"}]`
					}
					_, _ = fmt.Fprintf(w, `{"data":{"environments":{"nodes":%s}}}`, environments)
				case "GetTeamKafkaTopics":
					if tt.wantError != "" {
						t.Error("invalid environment reached a topic query")
					}
					_, _ = io.WriteString(w, `{"data":{"team":{"kafkaTopics":{"nodes":[]}}}}`)
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			app := newKafkaTestApplication(t, io.Discard)
			command := tt.command
			if command == nil {
				command = []string{"list", "-o", "json"}
			}
			args := append([]string{"kafka"}, command...)
			args = append(args, "-t", "my-team")
			if tt.environment != "" {
				args = append(args, "-e", tt.environment)
			}
			err := app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
