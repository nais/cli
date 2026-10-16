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

func TestMigratedJobCommandsUseNameChoices(t *testing.T) {
	for _, action := range []string{"activity", "issues", "log"} {
		t.Run(action, func(t *testing.T) {
			for _, tt := range []struct {
				name, value, wantError string
				completion, omitEnv    bool
			}{
				{name: "explicit valid name", value: "batch"},
				{name: "explicit invalid name", value: "unknown", wantError: "invalid value"},
				{name: "completion", value: "ba", completion: true},
				{name: "without environment", value: "batch", omitEnv: true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					var output bytes.Buffer
					app, global, err := naistrix.NewApplication("test", "Test job choices", "test", naistrix.ApplicationWithWriter(&output))
					if err != nil {
						t.Fatal(err)
					}
					global.Config = filepath.Join(t.TempDir(), "config.yaml")
					additional := &flags.AdditionalFlags{}
					if err := app.AddGlobalFlags(additional); err != nil {
						t.Fatal(err)
					}
					root := Job(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
					i := slices.IndexFunc(root.SubCommands, func(c *naistrix.Command) bool { return c.Name == action })
					if i == -1 {
						t.Fatalf("subcommand %q not found", action)
					}
					command := root.SubCommands[i]
					if command.AutoCompleteFunc != nil || command.Args[0].ChoicesFunc == nil {
						t.Fatal("command must use automatic job-name choices")
					}
					// Test argument resolution without starting log streaming or
					// querying activity or issues. The shared lookup has separate coverage.
					lookups, runs := 0, 0
					command.Args[0].ChoicesFunc = func(_ context.Context, args *naistrix.Arguments) ([]string, error) {
						lookups++
						wantEnv := "dev"
						if tt.omitEnv {
							wantEnv = ""
						}
						if args.Len() != 0 || additional.Team != "my-team" || string(additional.Environment) != wantEnv {
							t.Fatalf("unexpected lookup prerequisites: args=%v, flags=%+v", args.All(), additional)
						}
						return []string{"batch", "other"}, nil
					}
					command.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
						runs++
						if args.Get("name") != "batch" {
							t.Errorf("job = %q, want batch", args.Get("name"))
						}
						return nil
					}
					if err := app.AddCommand(root); err != nil {
						t.Fatal(err)
					}
					args := []string{"job", action, "-t", "my-team"}
					if !tt.omitEnv {
						args = append(args, "-e", "dev")
					}
					if tt.completion {
						args = append([]string{"__complete"}, args...)
					}
					args = append(args, tt.value)
					wantError := tt.wantError
					if tt.omitEnv && action == "log" {
						wantError = "missing required environment"
					}
					err = app.Run(naistrix.RunWithArgs(args))
					if wantError != "" {
						if err == nil || !strings.Contains(err.Error(), wantError) {
							t.Fatalf("error = %v, want %q", err, wantError)
						}
					} else if err != nil {
						t.Fatal(err)
					}
					wantLookups, wantRuns := 1, 0
					if tt.omitEnv && action == "log" {
						wantLookups = 0
					}
					if !tt.completion && wantError == "" {
						wantRuns = 1
					}
					if lookups != wantLookups || runs != wantRuns {
						t.Errorf("lookups/runs = %d/%d, want %d/%d", lookups, runs, wantLookups, wantRuns)
					}
					if tt.completion {
						want := "batch\n_activeHelp_ " + command.Args[0].Prompt + "\n:4\n"
						if output.String() != want {
							t.Errorf("completion = %q, want %q", output.String(), want)
						}
					}
				})
			}
		})
	}
}

func TestJobNameChoices(t *testing.T) {
	for _, action := range []struct {
		path      []string
		operation string
		setEnv    bool
	}{
		{path: []string{"trigger"}, operation: "TriggerJob"},
		{path: []string{"run", "list"}, operation: "GetJobRuns"},
		{path: []string{"set", "env"}, operation: "SetJobEnv", setEnv: true},
	} {
		t.Run(strings.Join(action.path, " "), func(t *testing.T) {
			for _, tt := range []struct {
				name, job, environment, config, prefix, wantError string
				completion, afterName, omitTeam, empty, apiError  bool
				setEnvOnly, omitEnvVars, invalidEnvVar            bool
			}{
				{name: "existing job", job: "batch", environment: "dev"},
				{name: "default environment", job: "batch", config: "environment: dev\n"},
				{name: "explicit environment overrides default", job: "batch", environment: "dev", config: "environment: prod\n"},
				{name: "unknown job", job: "unknown", environment: "dev", wantError: "invalid value"},
				{name: "job in another environment", job: "prod-only", environment: "dev", wantError: "invalid value"},
				{name: "no available jobs", job: "batch", environment: "dev", empty: true, wantError: "no choices available"},
				{name: "API error preserved", job: "batch", environment: "dev", apiError: true, wantError: "unable to fetch job names: input: access denied"},
				{name: "missing team stops lookup", job: "batch", environment: "dev", omitTeam: true, wantError: "missing required team"},
				{name: "missing environment stops lookup", job: "batch", wantError: "missing required environment"},
				{name: "sorted completion in selected environment", completion: true, environment: "dev"},
				{name: "completion in default environment", completion: true, config: "environment: dev\n"},
				{name: "completion in another environment", completion: true, environment: "prod"},
				{name: "completion prefix", completion: true, environment: "dev", prefix: "ba"},
				{name: "completion reports no jobs", completion: true, environment: "dev", empty: true, wantError: "no choices available"},
				{name: "completion reports API error", completion: true, environment: "dev", apiError: true, wantError: "unable to fetch job names: input: access denied"},
				{name: "completion validates team before lookup", completion: true, environment: "dev", omitTeam: true, wantError: "missing required team"},
				{name: "completion validates environment before lookup", completion: true, wantError: "missing required environment"},
				{name: "no name lookup after job argument", completion: true, afterName: true, job: "batch", environment: "dev", prefix: "ANY_KEY="},
				{name: "environment variables remain required", job: "batch", environment: "dev", setEnvOnly: true, omitEnvVars: true, wantError: "Expected at least 2 arguments"},
				{name: "invalid environment syntax still rejected", job: "batch", environment: "dev", setEnvOnly: true, invalidEnvVar: true, wantError: "invalid environment variable argument"},
			} {
				if tt.setEnvOnly && !action.setEnv {
					continue
				}
				t.Run(tt.name, func(t *testing.T) {
					var operations []string
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var request struct {
							OperationName string `json:"operationName"`
							Variables     struct {
								Team, Name, RunName  string
								Env                  json.RawMessage
								EnvironmentVariables []struct {
									Name  string
									Value *string
								}
							} `json:"variables"`
						}
						if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
							t.Errorf("decoding request: %v", err)
						}
						operations = append(operations, request.OperationName)
						v := request.Variables
						if v.Team != "my-team" {
							t.Errorf("team = %q, want my-team", v.Team)
						}
						var response string
						switch request.OperationName {
						case "GetJobNames":
							nodes := `[{"name":"zeta","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"batch","teamEnvironment":{"environment":{"name":"prod"}}},{"name":"batch","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"dev-only","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"prod-only","teamEnvironment":{"environment":{"name":"prod"}}}]`
							if tt.empty {
								nodes = `[]`
							}
							response = fmt.Sprintf(`{"data":{"team":{"jobs":{"nodes":%s}}}}`, nodes)
							if tt.apiError {
								response = `{"errors":[{"message":"access denied"}]}`
							}
						case action.operation:
							if v.Name != "batch" {
								t.Errorf("job = %q, want batch", v.Name)
							}
							wantEnv := `"dev"`
							if action.operation == "GetJobRuns" {
								wantEnv = `["dev"]`
							}
							if string(v.Env) != wantEnv {
								t.Errorf("environment = %s, want %s", v.Env, wantEnv)
							}
							switch action.operation {
							case "TriggerJob":
								if v.RunName != "custom-run" {
									t.Errorf("run name = %q, want custom-run", v.RunName)
								}
								response = `{"data":{"triggerJob":{"job":{"name":"batch"},"jobRun":{"name":"custom-run"}}}}`
							case "GetJobRuns":
								response = `{"data":{"team":{"jobs":{"nodes":[{"runs":{"nodes":[]}}]}}}}`
							case "SetJobEnv":
								vars := v.EnvironmentVariables
								if len(vars) != 3 || vars[0].Name != "NEW_KEY" || vars[0].Value == nil || *vars[0].Value != "arbitrary value=with=equals" || vars[1].Name != "EMPTY_KEY" || vars[1].Value == nil || *vars[1].Value != "" || vars[2].Name != "OLD_KEY" || vars[2].Value != nil {
									t.Errorf("incorrect environment variable updates: %+v", vars)
								}
								response = `{"data":{"updateJob":{"job":{"name":"batch"}}}}`
							}
						default:
							t.Errorf("unexpected API operation: %s", request.OperationName)
						}
						w.Header().Set("Content-Type", "application/json")
						if _, err := fmt.Fprint(w, response); err != nil {
							t.Errorf("writing response: %v", err)
						}
					}))
					defer server.Close()
					t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
					t.Setenv("HOME", t.TempDir())
					var output bytes.Buffer
					app, global, err := naistrix.NewApplication("test", "Test job choices", "test", naistrix.ApplicationWithWriter(&output))
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
					root := Job(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
					command := root
					for _, name := range action.path {
						i := slices.IndexFunc(command.SubCommands, func(c *naistrix.Command) bool { return c.Name == name })
						if i == -1 {
							t.Fatalf("subcommand %q not found", name)
						}
						command = command.SubCommands[i]
					}
					if command.AutoCompleteFunc != nil || command.Args[0].ChoicesFunc == nil {
						t.Fatal("command must use automatic job-name choices")
					}
					if action.setEnv && (!command.Args[1].Repeatable || len(command.Args[1].Choices) != 0 || command.Args[1].ChoicesFunc != nil) {
						t.Fatal("environment variable arguments must remain repeatable and unrestricted")
					}
					if err := app.AddCommand(root); err != nil {
						t.Fatal(err)
					}
					args := append([]string{"job"}, action.path...)
					if !tt.omitTeam {
						args = append(args, "-t", "my-team")
					}
					if tt.environment != "" {
						args = append(args, "-e", tt.environment)
					}
					if action.operation == "TriggerJob" {
						args = append(args, "--run-name", "custom-run")
					}
					if tt.completion {
						args = append([]string{"__complete"}, args...)
						if tt.afterName {
							args = append(args, tt.job)
						}
						args = append(args, tt.prefix)
					} else {
						args = append(args, tt.job)
						if action.setEnv && !tt.omitEnvVars {
							if tt.invalidEnvVar {
								args = append(args, "INVALID")
							} else {
								args = append(args, "NEW_KEY=arbitrary value=with=equals", "EMPTY_KEY=", "OLD_KEY-")
							}
						}
					}
					err = app.Run(naistrix.RunWithArgs(args))
					if tt.completion {
						if err != nil {
							t.Fatal(err)
						}
						got := output.String()
						if tt.wantError != "" {
							if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, tt.wantError) || strings.Count(got, "\n") != 2 {
								t.Errorf("completion = %q, want one active help error containing %q", got, tt.wantError)
							}
						} else if tt.afterName {
							if got != ":4\n" {
								t.Errorf("completion = %q, want no suggestions", got)
							}
						} else {
							wantNames := "batch\ndev-only\nzeta\n"
							if tt.prefix != "" {
								wantNames = "batch\n"
							} else if tt.environment == "prod" {
								wantNames = "batch\nprod-only\n"
							}
							want := wantNames + "_activeHelp_ " + command.Args[0].Prompt + "\n:4\n"
							if got != want {
								t.Errorf("completion = %q, want %q", got, want)
							}
						}
						if !strings.HasSuffix(got, ":4\n") {
							t.Errorf("completion = %q, want NoFileComp", got)
						}
					} else if tt.wantError != "" {
						if err == nil || !strings.Contains(err.Error(), tt.wantError) {
							t.Errorf("error = %v, want %q", err, tt.wantError)
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						if action.operation != "GetJobRuns" && (!strings.Contains(output.String(), "batch") || !strings.Contains(output.String(), "dev")) {
							t.Errorf("incorrect command output: %q", output.String())
						}
					}
					var wantOperations []string
					if !tt.omitTeam && (tt.environment != "" || tt.config != "") && !tt.afterName && !tt.omitEnvVars {
						wantOperations = append(wantOperations, "GetJobNames")
						if !tt.completion && tt.wantError == "" {
							wantOperations = append(wantOperations, action.operation)
						}
					}
					if !slices.Equal(operations, wantOperations) {
						t.Errorf("API operations = %v, want %v", operations, wantOperations)
					}
				})
			}
		})
	}
}
