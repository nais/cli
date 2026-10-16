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
