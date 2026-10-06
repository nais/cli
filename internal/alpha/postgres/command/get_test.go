package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestPostgresGetCommand(t *testing.T) {
	const configured = `{"name":"orders","majorVersion":"18","highAvailability":true,"resources":{"cpu":"200m","memory":"1Gi","diskSize":"20Gi"},"labels":[{"key":"domain","value":"payments"}],"activeBranch":{"name":"main"},"desiredActiveBranch":"restored","branches":{"nodes":[{"name":"main","state":"AVAILABLE"},{"name":"restored","state":"PROGRESSING"}]}}`
	const unconfigured = `{"name":"orders","majorVersion":"18","highAvailability":false,"resources":{"cpu":null,"memory":null,"diskSize":null},"labels":[],"activeBranch":null,"desiredActiveBranch":null,"branches":{"nodes":[]}}`
	const activated = `{"name":"orders","majorVersion":"18","highAvailability":false,"resources":{},"labels":[],"activeBranch":{"name":"main"},"desiredActiveBranch":"main","branches":{"nodes":[{"name":"main","state":"AVAILABLE"}]}}`
	for _, tt := range []struct {
		name, environment, config, format, metadata, lookupError, getError, wantError string
		lookupEnvironments                                                            []string
		complete, omitTeam                                                            bool
		wantOperations                                                                []string
	}{
		{name: "configured metadata as text", environment: "dev-gcp", metadata: configured, wantOperations: []string{"GetPostgres"}},
		{name: "configured metadata as JSON", environment: "dev-gcp", format: "json", metadata: configured, wantOperations: []string{"GetPostgres"}},
		{name: "unset requests are not inferred in text", environment: "dev-gcp", metadata: unconfigured, wantOperations: []string{"GetPostgres"}},
		{name: "unset requests remain null in JSON", environment: "dev-gcp", format: "json", metadata: unconfigured, wantOperations: []string{"GetPostgres"}},
		{name: "observed activation omits requested branch", environment: "dev-gcp", format: "json", metadata: activated, wantOperations: []string{"GetPostgres"}},
		{name: "observed activation text", environment: "dev-gcp", metadata: activated, wantOperations: []string{"GetPostgres"}},
		{name: "default environment bypasses discovery", config: "environment: dev-gcp\n", format: "json", metadata: configured, wantOperations: []string{"GetPostgres"}},
		{name: "explicit environment overrides default", environment: "dev-gcp", config: "environment: prod-gcp\n", format: "json", metadata: configured, wantOperations: []string{"GetPostgres"}},
		{name: "environment resolved from existing Postgres", lookupEnvironments: []string{"dev-gcp"}, format: "json", metadata: configured, wantOperations: []string{"GetTeamPostgreses", "GetPostgres"}},
		{name: "ambiguous environment requires choice", lookupEnvironments: []string{"dev-gcp", "prod-gcp"}, wantError: "specify environment with -e, --environment", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "missing Postgres", wantError: `Postgres "orders" not found`, wantOperations: []string{"GetTeamPostgreses"}},
		{name: "discovery error", lookupError: "access denied", wantError: "access denied", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "get error", environment: "dev-gcp", getError: "Resource not found", wantError: "Resource not found", wantOperations: []string{"GetPostgres"}},
		{name: "team is required", omitTeam: true, wantError: "missing required team"},
		{name: "registered completion includes branchless Postgres", complete: true, lookupEnvironments: []string{"dev-gcp"}, wantOperations: []string{"GetTeamPostgreses"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdin, err := os.Open(os.DevNull)
			if err != nil {
				t.Fatal(err)
			}
			previous := os.Stdin
			os.Stdin = stdin
			t.Cleanup(func() {
				os.Stdin = previous
				if err := stdin.Close(); err != nil {
					t.Error(err)
				}
			})
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string                                       `json:"operationName"`
					Query         string                                       `json:"query"`
					Variables     struct{ Team, Environment, Postgres string } `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetTeamPostgreses":
					if request.Variables.Team != "my-team" {
						t.Errorf("incorrect discovery team: %+v", request.Variables)
					}
					nodes := make([]string, 0, len(tt.lookupEnvironments))
					for _, env := range tt.lookupEnvironments {
						nodes = append(nodes, fmt.Sprintf(`{"name":"orders","teamEnvironment":{"environment":{"name":%q}},"activeBranch":null}`, env))
					}
					response = `{"data":{"team":{"postgreses":{"nodes":[` + strings.Join(nodes, ",") + `],"pageInfo":{"hasNextPage":false}}}}}`
					if tt.lookupError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.lookupError)
					}
				case "GetPostgres":
					if request.Variables.Team != "my-team" || request.Variables.Environment != "dev-gcp" || request.Variables.Postgres != "orders" {
						t.Errorf("incorrect metadata target: %+v", request.Variables)
					}
					for _, sensitive := range []string{"password", "relayToken", "postgresAccess"} {
						if strings.Contains(request.Query, sensitive) {
							t.Errorf("metadata query requests sensitive access field %s", sensitive)
						}
					}
					response = `{"data":{"team":{"environment":{"postgres":` + tt.metadata + `}}}}`
					if tt.getError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.getError)
					}
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(response)); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			config := filepath.Join(t.TempDir(), "config.yaml")
			contents := tt.config
			if contents == "" {
				contents = "{}\n"
			}
			if err := os.WriteFile(config, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("postgres-get-test", "Test Postgres metadata", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Postgres(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "--no-colors"}
			if !tt.omitTeam {
				args = append(args, "-t", "my-team")
			}
			if tt.complete {
				args = append(args, "__complete", "postgres", "get", "")
			} else {
				args = append(args, "postgres", "get", "orders")
				if tt.environment != "" {
					args = append(args, "-e", tt.environment)
				}
				if tt.format != "" {
					args = append(args, "--output", tt.format)
				}
			}
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v; want %q", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				switch {
				case tt.complete:
					if !strings.Contains(output.String(), "orders\n") {
						t.Errorf("missing name completion: %s", output.String())
					}
				case tt.format == "json":
					var detail postgresDetails
					if err := json.Unmarshal(output.Bytes(), &detail); err != nil {
						t.Fatalf("invalid JSON %q: %v", output.String(), err)
					}
					if detail.Name != "orders" || detail.Team != "my-team" || detail.Environment != "dev-gcp" || detail.Version != "18" {
						t.Errorf("incorrect metadata: %+v", detail)
					}
					if tt.metadata == configured {
						if !detail.HighAvailability || detail.Resources.CPU == nil || *detail.Resources.CPU != "200m" || detail.Resources.Memory == nil || *detail.Resources.Memory != "1Gi" || detail.Resources.DiskSize == nil || *detail.Resources.DiskSize != "20Gi" || len(detail.Labels) != 1 || detail.Labels[0].Key != "domain" || detail.Labels[0].Value != "payments" || detail.ActiveBranch != "main" || detail.RequestedBranch != "restored" || len(detail.Branches) != 2 {
							t.Fatalf("incorrect configuration detail: %+v", detail)
						}
						if detail.Branches[0].Name != "main" || !detail.Branches[0].Active || detail.Branches[0].Requested || detail.Branches[1].Name != "restored" || detail.Branches[1].Active || !detail.Branches[1].Requested {
							t.Errorf("incorrect branch markers: %+v", detail.Branches)
						}
					} else {
						if detail.HighAvailability || detail.Resources.CPU != nil || detail.Resources.Memory != nil || detail.Resources.DiskSize != nil || len(detail.Labels) != 0 || detail.RequestedBranch != "" {
							t.Errorf("inferred omitted configuration: %+v", detail)
						}
						if tt.metadata == unconfigured && (detail.ActiveBranch != "" || len(detail.Branches) != 0) {
							t.Errorf("invented branch observations: %+v", detail)
						}
						if strings.Contains(output.String(), `"requestedBranch"`) {
							t.Errorf("unexpected requested activation: %s", output.String())
						}
					}
					if strings.Contains(output.String(), `"status"`) {
						t.Errorf("get must not infer readiness: %s", output.String())
					}
				default:
					for _, field := range []string{"Name", "orders", "Team", "my-team", "Environment", "dev-gcp", "Version", "18", "High availability", "Requested CPU", "Requested memory", "Requested disk size", "Labels", "Branches", "Active branch"} {
						if !strings.Contains(output.String(), field) {
							t.Errorf("missing detail %s: %s", field, output.String())
						}
					}
					if tt.metadata == configured {
						for _, value := range []string{"200m", "1Gi", "20Gi", "true", "domain", "payments", "main", "restored", "activation pending", "Available", "Progressing"} {
							if !strings.Contains(output.String(), value) {
								t.Errorf("missing configuration/branches %s: %s", value, output.String())
							}
						}
						if strings.Count(output.String(), "Yes") != 2 {
							t.Errorf("missing active/requested branch table markers: %s", output.String())
						}
					} else {
						if strings.Count(output.String(), "(not configured)") != 3 || !strings.Contains(output.String(), "No labels configured.") || strings.Contains(output.String(), "activation pending") {
							t.Errorf("incorrect omitted configuration display: %s", output.String())
						}
						if tt.metadata == unconfigured && (!strings.Contains(output.String(), "(none)") || !strings.Contains(output.String(), "No Postgres branches found.")) {
							t.Errorf("incorrect absent branch display: %s", output.String())
						}
					}
					if strings.Contains(output.String(), "Status:") {
						t.Errorf("metadata command includes readiness summary: %s", output.String())
					}
				}
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
