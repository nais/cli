package application

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

	configcommand "github.com/nais/cli/internal/config/command"
	"github.com/nais/cli/internal/flags"
	secretcommand "github.com/nais/cli/internal/secret/command"
	"github.com/nais/naistrix"
)

func TestConfigAndSecretNameChoices(t *testing.T) {
	for _, resource := range []struct {
		name, operation, field string
		command                func(*flags.GlobalFlags) *naistrix.Command
	}{
		{name: "config", operation: "GetAllConfigs", field: "configs", command: configcommand.Config},
		{name: "secret", operation: "GetAllSecrets", field: "secrets", command: secretcommand.Secrets},
	} {
		for _, action := range []string{"activity", "delete", "set", "unset", "get"} {
			// Secret get also accepts platform-managed secrets absent from the list API.
			if resource.name == "secret" && action == "get" {
				continue
			}
			t.Run(resource.name+"/"+action, func(t *testing.T) {
				for _, tt := range []struct {
					name, value, environment, config, wantError      string
					completion, afterName, omitTeam, empty, apiError bool
					omitKey, omitValue                               bool
				}{
					{name: "existing name", value: "shared", environment: "dev"},
					{name: "default environment", value: "shared", config: "environment: dev\n"},
					{name: "explicit environment overrides default", value: "shared", environment: "dev", config: "environment: prod\n"},
					{name: "unknown name", value: "unknown", environment: "dev", wantError: "invalid value"},
					{name: "name in another environment", value: "prod-only", environment: "dev", wantError: "invalid value"},
					{name: "no resources", value: "shared", environment: "dev", empty: true, wantError: "no choices available"},
					{name: "API error", value: "shared", environment: "dev", apiError: true, wantError: "access denied"},
					{name: "missing team", value: "shared", environment: "dev", omitTeam: true, wantError: "missing required team"},
					{name: "missing environment", value: "shared", wantError: "missing required environment"},
					{name: "completion in environment", completion: true, environment: "dev"},
					{name: "completion in default environment", completion: true, config: "environment: dev\n"},
					{name: "completion in another environment", completion: true, environment: "prod"},
					{name: "completion prefix", completion: true, value: "sha", environment: "dev"},
					{name: "completion reports no resources", completion: true, environment: "dev", empty: true, wantError: "no choices available"},
					{name: "completion reports API error", completion: true, environment: "dev", apiError: true, wantError: "access denied"},
					{name: "completion requires team", completion: true, environment: "dev", omitTeam: true, wantError: "missing required team"},
					{name: "completion requires environment", completion: true, wantError: "missing required environment"},
					{name: "no completion after name", completion: true, afterName: true, environment: "dev"},
					{name: "key remains required", value: "shared", environment: "dev", omitKey: true, wantError: "--key is required"},
					{name: "completion requires key", completion: true, environment: "dev", omitKey: true, wantError: "--key is required"},
					{name: "value remains required", value: "shared", environment: "dev", omitValue: true, wantError: "--value, --value-from-stdin, or --value-from-file is required"},
					{name: "completion requires value", completion: true, environment: "dev", omitValue: true, wantError: "--value, --value-from-stdin, or --value-from-file is required"},
				} {
					if tt.omitKey && action != "set" && action != "unset" || tt.omitValue && action != "set" {
						continue
					}
					t.Run(tt.name, func(t *testing.T) {
						lookups := 0
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							var request struct {
								OperationName string `json:"operationName"`
								Variables     struct {
									TeamSlug string
									Filter   struct{ Environments []string }
								} `json:"variables"`
							}
							if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
								t.Errorf("decoding request: %v", err)
							}
							lookups++
							if request.OperationName != resource.operation || request.Variables.TeamSlug != "my-team" {
								t.Errorf("unexpected name lookup: %+v", request)
							}
							wantEnv := tt.environment
							if wantEnv == "" {
								wantEnv = "dev"
							}
							if !slices.Equal(request.Variables.Filter.Environments, []string{wantEnv}) {
								t.Errorf("environment filter = %v, want [%s]", request.Variables.Filter.Environments, wantEnv)
							}
							// Names are unique within the environment selected by the API filter.
							nodes := `[{"name":"zeta"},{"name":"shared"}]`
							if slices.Contains(request.Variables.Filter.Environments, "prod") {
								nodes = `[{"name":"shared"},{"name":"prod-only"}]`
							}
							if tt.empty {
								nodes = `[]`
							}
							response := fmt.Sprintf(`{"data":{"team":{%q:{"nodes":%s}}}}`, resource.field, nodes)
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
						app, global, err := naistrix.NewApplication("test", "Test resource choices", "test", naistrix.ApplicationWithWriter(&output))
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
						root := resource.command(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
						i := slices.IndexFunc(root.SubCommands, func(c *naistrix.Command) bool { return c.Name == action })
						if i == -1 {
							t.Fatalf("subcommand %q not found", action)
						}
						command := root.SubCommands[i]
						if command.AutoCompleteFunc != nil || command.Args[0].ChoicesFunc == nil {
							t.Fatal("command must use automatic resource-name choices")
						}
						for _, child := range root.SubCommands {
							if child.Name == "create" {
								if child.Args[0].ChoicesFunc != nil || child.AutoCompleteFunc != nil {
									t.Fatalf("%s must retain its existing argument and completion contract", child.Name)
								}
							}
						}
						// Exercise argument resolution without mutations or confirmation prompts.
						runs := 0
						command.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
							runs++
							if args.Get("name") != tt.value {
								t.Errorf("name = %q, want %q", args.Get("name"), tt.value)
							}
							return nil
						}
						if err := app.AddCommand(root); err != nil {
							t.Fatal(err)
						}
						args := []string{resource.name, action}
						if !tt.omitTeam {
							args = append(args, "-t", "my-team")
						}
						if tt.environment != "" {
							args = append(args, "-e", tt.environment)
						}
						if (action == "set" || action == "unset") && !tt.omitKey {
							args = append(args, "--key", "SOME_KEY")
						}
						if action == "set" && !tt.omitValue {
							args = append(args, "--value", "arbitrary value=with=equals")
						}
						if tt.completion {
							args = append([]string{"__complete"}, args...)
							if tt.afterName {
								args = append(args, "shared")
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
						if tt.omitTeam || tt.environment == "" && tt.config == "" || tt.afterName || tt.omitKey || tt.omitValue {
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
									names := "shared\nzeta\n"
									if tt.value != "" {
										names = "shared\n"
									} else if tt.environment == "prod" {
										names = "prod-only\nshared\n"
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
}
