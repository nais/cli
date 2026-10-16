package command

import (
	"bytes"
	"context"
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

func TestPostgresNameChoices(t *testing.T) {
	for _, path := range [][]string{{"get"}, {"update"}, {"delete"}, {"psql"}, {"proxy"}, {"branch", "list"}} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			for _, tt := range []struct {
				name, value, environment, config, wantError      string
				completion, afterName, omitTeam, empty, apiError bool
			}{
				{name: "existing instance", value: "orders", environment: "dev-gcp"},
				{name: "branchless instance", value: "branchless", environment: "dev-gcp"},
				{name: "default environment", value: "orders", config: "environment: dev-gcp\n"},
				{name: "explicit environment overrides default", value: "orders", environment: "dev-gcp", config: "environment: prod-gcp\n"},
				{name: "without environment", value: "orders"},
				{name: "unknown instance", value: "unknown", wantError: "invalid value"},
				{name: "instance in another environment", value: "prod-only", environment: "dev-gcp", wantError: "invalid value"},
				{name: "no instances", value: "orders", empty: true, wantError: "no choices available"},
				{name: "API error", value: "orders", apiError: true, wantError: "unable to fetch Postgres names: input: access denied"},
				{name: "missing team", value: "orders", omitTeam: true, wantError: "missing required team"},
				{name: "completion across environments", completion: true},
				{name: "completion in environment", completion: true, environment: "dev-gcp"},
				{name: "completion in default environment", completion: true, config: "environment: dev-gcp\n"},
				{name: "completion prefix", completion: true, value: "ord"},
				{name: "completion reports empty list", completion: true, empty: true, wantError: "no choices available"},
				{name: "completion reports API error", completion: true, apiError: true, wantError: "unable to fetch Postgres names: input: access denied"},
				{name: "completion validates team", completion: true, omitTeam: true, wantError: "missing required team"},
				{name: "no completion after name", completion: true, afterName: true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					lookups := 0
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var request struct {
							OperationName string                `json:"operationName"`
							Variables     struct{ Team string } `json:"variables"`
						}
						if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
							t.Errorf("decoding request: %v", err)
						}
						lookups++
						if request.OperationName != "GetTeamPostgreses" || request.Variables.Team != "my-team" {
							t.Errorf("unexpected name lookup: %+v", request)
						}
						nodes := `[{"name":"orders","teamEnvironment":{"environment":{"name":"prod-gcp"}},"activeBranch":null},{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}},"activeBranch":null},{"name":"branchless","teamEnvironment":{"environment":{"name":"dev-gcp"}},"activeBranch":null},{"name":"prod-only","teamEnvironment":{"environment":{"name":"prod-gcp"}},"activeBranch":null}]`
						if tt.empty {
							nodes = `[]`
						}
						response := fmt.Sprintf(`{"data":{"team":{"postgreses":{"nodes":%s,"pageInfo":{"hasNextPage":false}}}}}`, nodes)
						if tt.apiError {
							response = `{"errors":[{"message":"access denied"}]}`
						}
						w.Header().Set("Content-Type", "application/json")
						if _, err := fmt.Fprint(w, response); err != nil {
							t.Errorf("writing response: %v", err)
						}
					}))
					defer server.Close()
					t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
					t.Setenv("HOME", t.TempDir())
					t.Setenv("NAIS_TEAM", "")
					t.Setenv("NAIS_ENVIRONMENT", "")
					var output bytes.Buffer
					app, global, err := naistrix.NewApplication("test", "Test Postgres choices", "test", naistrix.ApplicationWithWriter(&output))
					if err != nil {
						t.Fatal(err)
					}
					global.Config = filepath.Join(t.TempDir(), "config.yaml")
					if tt.config != "" {
						if err := os.WriteFile(global.Config, []byte(tt.config), 0o600); err != nil {
							t.Fatal(err)
						}
					}
					additional := &flags.AdditionalFlags{}
					if err := app.AddGlobalFlags(additional); err != nil {
						t.Fatal(err)
					}
					root := Postgres(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
					command := root
					for _, name := range path {
						i := slices.IndexFunc(command.SubCommands, func(c *naistrix.Command) bool { return c.Name == name })
						if i == -1 {
							t.Fatalf("subcommand %q not found", name)
						}
						command = command.SubCommands[i]
					}
					if command.AutoCompleteFunc != nil || command.Args[0].ChoicesFunc == nil {
						t.Fatal("command must use automatic Postgres-name choices")
					}
					// Exercise name resolution without mutations, prompts, or connections.
					runs := 0
					command.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
						runs++
						if args.Get("postgres") != tt.value {
							t.Errorf("Postgres = %q, want %q", args.Get("postgres"), tt.value)
						}
						return nil
					}
					if err := app.AddCommand(root); err != nil {
						t.Fatal(err)
					}
					args := append([]string{"postgres"}, path...)
					if !tt.omitTeam {
						args = append(args, "-t", "my-team")
					}
					if tt.environment != "" {
						args = append(args, "-e", tt.environment)
					}
					if tt.completion {
						args = append([]string{"__complete"}, args...)
						if tt.afterName {
							args = append(args, "orders")
						}
					}
					args = append(args, tt.value)
					err = app.Run(naistrix.RunWithArgs(args))
					if tt.wantError != "" && !tt.completion {
						if err == nil || !strings.Contains(err.Error(), tt.wantError) {
							t.Fatalf("error = %v, want %q", err, tt.wantError)
						}
					} else if err != nil {
						t.Fatal(err)
					}
					wantLookups, wantRuns := 1, 0
					if tt.omitTeam || tt.afterName {
						wantLookups = 0
					}
					if !tt.completion && tt.wantError == "" {
						wantRuns = 1
					}
					if lookups != wantLookups || runs != wantRuns {
						t.Errorf("lookups/runs = %d/%d, want %d/%d", lookups, runs, wantLookups, wantRuns)
					}
					if tt.completion {
						got := output.String()
						if tt.wantError != "" {
							if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, tt.wantError) || strings.Count(got, "\n") != 2 || !strings.HasSuffix(got, ":4\n") {
								t.Errorf("completion = %q, want one error containing %q and NoFileComp", got, tt.wantError)
							}
						} else {
							want := ":4\n"
							if !tt.afterName {
								names := "branchless\norders\nprod-only\n"
								if tt.value != "" {
									names = "orders\n"
								} else if tt.environment != "" || tt.config != "" {
									names = "branchless\norders\n"
								}
								want = names + "_activeHelp_ " + command.Args[0].Prompt + "\n:4\n"
							}
							if got != want {
								t.Errorf("completion = %q, want %q", got, want)
							}
						}
					}
				})
			}
		})
	}
}
