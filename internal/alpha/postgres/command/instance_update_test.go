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

func TestPostgresUpdateComparison(t *testing.T) {
	const configured = `{"name":"orders","majorVersion":"18","highAvailability":true,"resources":{"cpu":"100m","memory":"512Mi","diskSize":"10Gi"},"labels":[],"branches":{"nodes":[]}}`
	const unconfigured = `{"name":"orders","majorVersion":"18","highAvailability":false,"resources":{"cpu":null,"memory":null,"diskSize":null},"labels":[],"branches":{"nodes":[]}}`
	for _, tt := range []struct {
		name, metadata, config, lookupError, discoveryError, wantError string
		args                                                           []string
		resolveEnvironment, omitTeam, omitName, confirm                bool
		wantRows                                                       []string
		wantChanges                                                    string
		wantOperations                                                 []string
	}{
		{
			name: "all provided configured values including false HA", metadata: configured,
			args:        []string{"--high-availability", "false", "--cpu", "200m", "--memory", "1Gi", "--disk-size", "20Gi"},
			wantRows:    []string{"High availability true false", "CPU 100m 200m", "Memory 512Mi 1Gi", "Disk size 10Gi 20Gi"},
			wantChanges: `{"highAvailability":false,"cpu":"200m","memory":"1Gi","diskSize":"20Gi"}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "false configured HA can be enabled", metadata: unconfigured, args: []string{"--high-availability", "true"},
			wantRows: []string{"High availability false true"}, wantChanges: `{"highAvailability":true}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "omitted resource requests are not inferred", metadata: unconfigured,
			args:        []string{"--cpu", "200m", "--memory", "1Gi", "--disk-size", "20Gi"},
			wantRows:    []string{"CPU (not configured) 200m", "Memory (not configured) 1Gi", "Disk size (not configured) 20Gi"},
			wantChanges: `{"cpu":"200m","memory":"1Gi","diskSize":"20Gi"}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "partial request excludes untouched settings", metadata: configured, args: []string{"--cpu", "200m"},
			wantRows: []string{"CPU 100m 200m"}, wantChanges: `{"cpu":"200m"}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "same configured values remain requested", metadata: configured,
			args:        []string{"--high-availability", "true", "--cpu", "100m", "--memory", "512Mi", "--disk-size", "10Gi"},
			wantRows:    []string{"High availability true true", "CPU 100m 100m", "Memory 512Mi 512Mi", "Disk size 10Gi 10Gi"},
			wantChanges: `{"highAvailability":true,"cpu":"100m","memory":"512Mi","diskSize":"10Gi"}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "default environment bypasses discovery", config: "environment: dev-gcp\n", resolveEnvironment: true, metadata: configured,
			args: []string{"--memory", "1Gi"}, wantRows: []string{"Memory 512Mi 1Gi"}, wantChanges: `{"memory":"1Gi"}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "explicit environment overrides default", config: "environment: prod-gcp\n", metadata: configured,
			args: []string{"--disk-size", "20Gi"}, wantRows: []string{"Disk size 10Gi 20Gi"}, wantChanges: `{"diskSize":"20Gi"}`, wantOperations: []string{"GetPostgres", "UpdatePostgres"},
		},
		{
			name: "environment resolved from existing Postgres", resolveEnvironment: true, metadata: configured,
			args: []string{"--cpu", "200m"}, wantRows: []string{"CPU 100m 200m"}, wantChanges: `{"cpu":"200m"}`, wantOperations: []string{"GetTeamPostgreses", "GetPostgres", "UpdatePostgres"},
		},
		{
			name: "lookup failure stops before confirmation", args: []string{"--cpu", "200m"}, lookupError: "access denied", confirm: true,
			wantError: "access denied", wantOperations: []string{"GetPostgres"},
		},
		{
			name: "discovery failure stops before metadata lookup", args: []string{"--cpu", "200m"}, resolveEnvironment: true, discoveryError: "access denied",
			wantError: "access denied", wantOperations: []string{"GetTeamPostgreses"},
		},
		{
			name: "confirmation unavailable does not mutate", metadata: configured, args: []string{"--cpu", "200m"}, confirm: true,
			wantRows: []string{"CPU 100m 200m"}, wantError: "no interactive terminal available", wantOperations: []string{"GetPostgres"},
		},
		{name: "team required before API calls", omitTeam: true, args: []string{"--cpu", "200m"}, wantError: "missing required team"},
		{name: "name required before API calls", omitName: true, args: []string{"--cpu", "200m"}, wantError: "Expected exactly 1 argument"},
		{name: "at least one field required before discovery", resolveEnvironment: true, wantError: "specify at least one of"},
		{name: "invalid HA rejected before discovery", resolveEnvironment: true, args: []string{"--high-availability", "yes"}, wantError: "--high-availability must be true or false"},
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
			var updated map[string]json.RawMessage
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Team, Environment, Postgres string
						Input                       map[string]json.RawMessage `json:"input"`
					} `json:"variables"`
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
					response = `{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}},"activeBranch":null}],"pageInfo":{"hasNextPage":false}}}}}`
					if tt.discoveryError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.discoveryError)
					}
				case "GetPostgres":
					if request.Variables.Team != "my-team" || request.Variables.Environment != "dev-gcp" || request.Variables.Postgres != "orders" {
						t.Errorf("incorrect configured snapshot target: %+v", request.Variables)
					}
					response = `{"data":{"team":{"environment":{"postgres":` + tt.metadata + `}}}}`
					if tt.lookupError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.lookupError)
					}
				case "UpdatePostgres":
					updated = request.Variables.Input
					response = `{"data":{"updatePostgres":{"postgres":{"name":"orders"}}}}`
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
			app, global, err := naistrix.NewApplication("postgres-update-test", "Test Postgres update", "test", naistrix.ApplicationWithWriter(&output))
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
			args := []string{"--config", config, "--no-colors", "postgres", "update"}
			if !tt.omitName {
				args = append(args, "orders")
			}
			if !tt.omitTeam {
				args = append(args, "-t", "my-team")
			}
			if !tt.resolveEnvironment {
				args = append(args, "-e", "dev-gcp")
			}
			if !tt.confirm {
				args = append(args, "--yes")
			}
			args = append(args, tt.args...)
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v; want %q", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				var expected map[string]json.RawMessage
				if err := json.Unmarshal([]byte(tt.wantChanges), &expected); err != nil {
					t.Fatal(err)
				}
				expected["name"], expected["teamSlug"], expected["environmentName"] = json.RawMessage(`"orders"`), json.RawMessage(`"my-team"`), json.RawMessage(`"dev-gcp"`)
				if len(updated) != len(expected) {
					t.Errorf("mutation fields = %v; want %v", updated, expected)
				}
				for field, want := range expected {
					if string(updated[field]) != string(want) {
						t.Errorf("mutation field %s = %s; want %s", field, updated[field], want)
					}
				}
				if !strings.Contains(output.String(), `Postgres "orders" update requested in "dev-gcp"; reconciliation may still be in progress.`) {
					t.Errorf("missing asynchronous result: %s", output.String())
				}
			}
			if len(tt.wantRows) > 0 {
				var rows []string
				for _, line := range strings.Split(output.String(), "\n") {
					line = strings.NewReplacer("|", " ", "│", " ").Replace(line)
					rows = append(rows, strings.Join(strings.Fields(line), " "))
				}
				for _, want := range append([]string{"Setting Current Requested", "Team my-team", "Environment dev-gcp", "Name orders"}, tt.wantRows...) {
					if !slices.Contains(rows, want) {
						t.Errorf("missing comparison/target row %q: %s", want, output.String())
					}
				}
				for _, label := range []string{"High availability", "CPU", "Memory", "Disk size"} {
					requested := false
					for _, row := range tt.wantRows {
						requested = requested || strings.HasPrefix(row, label+" ")
					}
					if !requested {
						for _, row := range rows {
							if strings.HasPrefix(row, label+" ") {
								t.Errorf("untouched setting %q is displayed: %s", label, output.String())
							}
						}
					}
				}
				if !strings.Contains(output.String(), "can be overwritten by a subsequent nais apply") || !strings.Contains(output.String(), "configured snapshot") || !strings.Contains(output.String(), "effective runtime resources or SQL readiness") {
					t.Errorf("missing manifest/configuration scope warning: %s", output.String())
				}
			} else if strings.Contains(output.String(), "Setting") || strings.Contains(output.String(), "subsequent nais apply") {
				t.Errorf("summary appeared before successful metadata lookup: %s", output.String())
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
