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

func TestMigratedApplicationCommandsUseNameChoices(t *testing.T) {
	for _, path := range [][]string{
		{"activity"},
		{"env"},
		{"files"},
		{"issues"},
		{"log"},
		{"delete"},
		{"stop"},
		{"set", "env"},
		{"set", "image"},
		{"set", "replicas"},
		{"labels", "list"},
		{"labels", "set"},
	} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			for _, tt := range []struct {
				name, value, wantError            string
				completion, afterName, omitLabels bool
				replicaFlags                      []string
			}{
				{name: "explicit valid name", value: "orders"},
				{name: "explicit invalid name", value: "unknown", wantError: "invalid value"},
				{name: "completion", completion: true, value: "ord"},
				{name: "no completion after name", completion: true, afterName: true},
				{name: "labels remain required", value: "orders", omitLabels: true, wantError: "at least one --label KEY=VALUE must be specified"},
				{name: "completion requires labels", completion: true, omitLabels: true, wantError: "at least one --label KEY=VALUE must be specified"},
				{name: "negative minimum rejected", value: "orders", replicaFlags: []string{"--min=-1", "--max=5"}, wantError: "--min must be a non-negative integer"},
				{name: "negative maximum rejected", value: "orders", replicaFlags: []string{"--min=0", "--max=-1"}, wantError: "--max must be a non-negative integer"},
				{name: "inverted range rejected", value: "orders", replicaFlags: []string{"--min=5", "--max=2"}, wantError: "--min (5) cannot be greater than --max (2)"},
				{name: "completion validates replica range", completion: true, replicaFlags: []string{"--min=5", "--max=2"}, wantError: "--min (5) cannot be greater than --max (2)"},
			} {
				setLabels := slices.Equal(path, []string{"labels", "set"})
				setReplicas := slices.Equal(path, []string{"set", "replicas"})
				if tt.omitLabels && !setLabels {
					continue
				}
				if tt.replicaFlags != nil && !setReplicas {
					continue
				}
				t.Run(tt.name, func(t *testing.T) {
					var output bytes.Buffer
					app, global, err := naistrix.NewApplication("test", "Test application choices", "test", naistrix.ApplicationWithWriter(&output))
					if err != nil {
						t.Fatal(err)
					}
					global.Config = filepath.Join(t.TempDir(), "config.yaml")
					additional := &flags.AdditionalFlags{}
					if err := app.AddGlobalFlags(additional); err != nil {
						t.Fatal(err)
					}
					root := App(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
					command := root
					for _, name := range path {
						i := slices.IndexFunc(command.SubCommands, func(c *naistrix.Command) bool { return c.Name == name })
						if i == -1 {
							t.Fatalf("subcommand %q not found", name)
						}
						command = command.SubCommands[i]
					}
					if command.AutoCompleteFunc != nil || command.Args[0].ChoicesFunc == nil {
						t.Fatal("command must use automatic application-name choices")
					}
					setEnv := slices.Equal(path, []string{"set", "env"})
					if setEnv && (!command.Args[1].Repeatable || len(command.Args[1].Choices) != 0 || command.Args[1].ChoicesFunc != nil) {
						t.Fatal("environment variable arguments must remain repeatable and unrestricted")
					}
					// The shared lookup has separate coverage. Exercise each command's argument
					// wiring without running mutations, interactive prompts, or log streaming.
					lookups, runs := 0, 0
					command.Args[0].ChoicesFunc = func(_ context.Context, args *naistrix.Arguments) ([]string, error) {
						lookups++
						if args.Len() != 0 || additional.Team != "my-team" || additional.Environment != "dev" {
							t.Fatalf("unexpected lookup prerequisites: args=%v, flags=%+v", args.All(), additional)
						}
						return []string{"orders", "other"}, nil
					}
					vars := []string{"NEW_KEY=arbitrary value=with=equals", "EMPTY_KEY=", "OLD_KEY-"}
					command.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
						runs++
						if args.Get("name") != "orders" {
							t.Errorf("application = %q, want orders", args.Get("name"))
						}
						if setEnv && !slices.Equal(args.GetRepeatable("env_vars"), vars) {
							t.Errorf("environment variables = %v, want %v", args.GetRepeatable("env_vars"), vars)
						}
						return nil
					}
					if err := app.AddCommand(root); err != nil {
						t.Fatal(err)
					}
					args := append([]string{"app"}, path...)
					args = append(args, "-t", "my-team", "-e", "dev")
					if setLabels && !tt.omitLabels {
						args = append(args, "--label", "foo=bar")
					}
					if setReplicas {
						if tt.replicaFlags != nil {
							args = append(args, tt.replicaFlags...)
						} else {
							args = append(args, "--min=2", "--max=5")
						}
					}
					if tt.completion {
						args = append([]string{"__complete"}, args...)
						if tt.afterName {
							args = append(args, "orders")
						}
						args = append(args, tt.value)
					} else {
						args = append(args, tt.value)
						if setEnv {
							args = append(args, vars...)
						}
					}
					err = app.Run(naistrix.RunWithArgs(args))
					if tt.wantError != "" && !tt.completion {
						if err == nil || !strings.Contains(err.Error(), tt.wantError) {
							t.Fatalf("error = %v, want %q", err, tt.wantError)
						}
					} else if err != nil {
						t.Fatal(err)
					}
					wantLookups, wantRuns := 1, 0
					if tt.afterName || tt.omitLabels || tt.replicaFlags != nil {
						wantLookups = 0
					}
					if !tt.completion && tt.wantError == "" {
						wantRuns = 1
					}
					if lookups != wantLookups || runs != wantRuns {
						t.Errorf("lookups/runs = %d/%d, want %d/%d", lookups, runs, wantLookups, wantRuns)
					}
					if tt.completion {
						want := ":4\n"
						if tt.wantError != "" {
							want = "_activeHelp_ Completion failed: " + tt.wantError + "\n:4\n"
						} else if !tt.afterName {
							want = "orders\n_activeHelp_ " + command.Args[0].Prompt + "\n:4\n"
						}
						if output.String() != want {
							t.Errorf("completion = %q, want %q", output.String(), want)
						}
					}
				})
			}
		})
	}
}

func TestApplicationNameChoices(t *testing.T) {
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
	for _, action := range []string{"status", "restart"} {
		t.Run(action, func(t *testing.T) {
			for _, tt := range []struct {
				name, application, environment, config, prefix, wantError string
				completion, afterValue, omitTeam, empty, apiError         bool
			}{
				{name: "explicit environment", application: "orders", environment: "dev"},
				{name: "default environment", application: "orders", config: "environment: dev\n"},
				{name: "explicit environment overrides default", application: "orders", environment: "dev", config: "environment: prod\n"},
				{name: "unknown application", application: "unknown", environment: "dev", wantError: "invalid value"},
				{name: "application in another environment", application: "prod-only", environment: "dev", wantError: "invalid value"},
				{name: "no available applications", application: "orders", environment: "dev", empty: true, wantError: "no choices available"},
				{name: "API error preserved", application: "orders", environment: "dev", apiError: true, wantError: "unable to fetch application names: input: access denied"},
				{name: "missing team stops lookup", application: "orders", environment: "dev", omitTeam: true, wantError: "missing required team"},
				{name: "omitted environment still reaches environment resolution", application: "orders", wantError: "specify environment with -e, --environment"},
				{name: "completion across environments", completion: true},
				{name: "completion in selected environment", completion: true, environment: "dev"},
				{name: "completion in default environment", completion: true, config: "environment: dev\n"},
				{name: "completion in another environment", completion: true, environment: "prod"},
				{name: "completion prefix", completion: true, prefix: "ord"},
				{name: "completion reports empty list", completion: true, empty: true, wantError: "no choices available"},
				{name: "completion reports API error", completion: true, apiError: true, wantError: "unable to fetch application names: input: access denied"},
				{name: "completion validates team before lookup", completion: true, omitTeam: true, wantError: "missing required team"},
				{name: "no lookup after final argument", completion: true, afterValue: true, application: "orders"},
			} {
				t.Run(tt.name, func(t *testing.T) {
					var operations []string
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var request struct {
							OperationName string `json:"operationName"`
							Variables     struct {
								Team, Slug, Name, Application string
								Env                           json.RawMessage
								Filter                        struct{ Name string }
							} `json:"variables"`
						}
						if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
							t.Errorf("decoding request: %v", err)
						}
						operations = append(operations, request.OperationName)
						v := request.Variables
						if v.Team != "my-team" && v.Slug != "my-team" {
							t.Errorf("incorrect team: %+v", v)
						}
						var response string
						switch request.OperationName {
						case "GetApplicationNames":
							nodes := `[{"name":"zeta","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"orders","teamEnvironment":{"environment":{"name":"prod"}}},{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"dev-only","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"prod-only","teamEnvironment":{"environment":{"name":"prod"}}}]`
							if tt.empty {
								nodes = `[]`
							}
							response = fmt.Sprintf(`{"data":{"team":{"applications":{"nodes":%s}}}}`, nodes)
							if tt.apiError {
								response = `{"errors":[{"message":"access denied"}]}`
							}
						case "ApplicationEnvironments":
							if v.Filter.Name != "orders" {
								t.Errorf("environment lookup name = %q, want orders", v.Filter.Name)
							}
							response = `{"data":{"team":{"applications":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"prod"}}},{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}}}]}}}}`
						case "GetApplicationStatus":
							if action != "status" || v.Name != "orders" || string(v.Env) != `["dev"]` {
								t.Errorf("incorrect status target: %+v", v)
							}
							response = `{"data":{"team":{"applications":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}},"instanceGroups":[]}]}}}}`
						case "RestartApp":
							if action != "restart" || v.Application != "orders" || string(v.Env) != `"dev"` {
								t.Errorf("incorrect restart target: %+v", v)
							}
							response = `{"data":{"restartApplication":{"application":{"name":"orders"}}}}`
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
					app, global, err := naistrix.NewApplication("test", "Test application choices", "test", naistrix.ApplicationWithWriter(&output))
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
					root := App(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
					var command *naistrix.Command
					for _, child := range root.SubCommands {
						if child.Name == action {
							command = child
						}
					}
					if command == nil || command.AutoCompleteFunc != nil || command.Args[0].ChoicesFunc == nil {
						t.Fatal("command must use automatic application-name choices")
					}
					if err := app.AddCommand(root); err != nil {
						t.Fatal(err)
					}
					args := []string{"app", action}
					if !tt.omitTeam {
						args = append(args, "-t", "my-team")
					}
					if tt.environment != "" {
						args = append(args, "-e", tt.environment)
					}
					if tt.completion {
						args = append([]string{"__complete"}, args...)
						if tt.afterValue {
							args = append(args, tt.application)
						}
						args = append(args, tt.prefix)
					} else {
						args = append(args, tt.application)
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
						} else if tt.afterValue {
							if got != ":4\n" {
								t.Errorf("completion = %q, want no suggestions", got)
							}
						} else {
							wantNames := "dev-only\norders\nprod-only\nzeta\n"
							switch {
							case tt.prefix != "":
								wantNames = "orders\n"
							case tt.environment == "dev" || tt.config != "":
								wantNames = "dev-only\norders\nzeta\n"
							case tt.environment == "prod":
								wantNames = "orders\nprod-only\n"
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
						if !strings.Contains(output.String(), "orders") || !strings.Contains(output.String(), "dev") {
							t.Errorf("incorrect command output: %q", output.String())
						}
					}
					var wantOperations []string
					if !tt.omitTeam && !tt.afterValue {
						wantOperations = append(wantOperations, "GetApplicationNames")
						if !tt.completion && !tt.apiError && !tt.empty && tt.application == "orders" {
							wantOperations = append(wantOperations, "ApplicationEnvironments")
							if tt.wantError == "" {
								operation := "GetApplicationStatus"
								if action == "restart" {
									operation = "RestartApp"
								}
								wantOperations = append(wantOperations, operation)
							}
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
