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

func TestExistingInstanceChoices(t *testing.T) {
	for _, action := range []string{"get", "update", "delete", "credentials", "create"} {
		t.Run(action, func(t *testing.T) {
			for _, tt := range []struct {
				name, value, prefix, wantError                string
				completion, afterValue, empty, apiError       bool
				omitTeam, omitEnvironment, defaultEnvironment bool
			}{
				{name: "existing instance", value: "alpha"},
				{name: "default environment", value: "alpha", defaultEnvironment: true},
				{name: "unknown instance", value: "unknown", wantError: "invalid value"},
				{name: "instance from another environment", value: "prod-only", wantError: "invalid value"},
				{name: "empty list", value: "alpha", empty: true, wantError: "no choices available"},
				{name: "API error", value: "alpha", apiError: true, wantError: "access denied"},
				{name: "missing team", value: "alpha", omitTeam: true, wantError: "missing required team"},
				{name: "missing environment", value: "alpha", omitEnvironment: true, wantError: "missing required environment"},
				{name: "sorted completion", completion: true},
				{name: "completion prefix", completion: true, prefix: "sh"},
				{name: "completion API error", completion: true, apiError: true, wantError: "access denied"},
				{name: "completion empty list", completion: true, empty: true, wantError: "no choices available"},
				{name: "completion missing team", completion: true, omitTeam: true, wantError: "missing required team"},
				{name: "completion missing environment", completion: true, omitEnvironment: true, wantError: "missing required environment"},
				{name: "completion default environment", completion: true, defaultEnvironment: true},
				{name: "no lookup after final argument", value: "alpha", completion: true, afterValue: true},
			} {
				if action == "create" && tt.name != "unknown instance" && tt.name != "sorted completion" {
					continue
				}
				t.Run(tt.name, func(t *testing.T) {
					var operations []string
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
						operations = append(operations, request.OperationName)
						if request.Variables.TeamSlug != "my-team" || !slices.Equal(request.Variables.Filter.Environments, []string{"dev"}) {
							t.Errorf("incorrect instance lookup: %+v", request.Variables)
						}
						w.Header().Set("Content-Type", "application/json")
						var response string
						if tt.apiError {
							response = `{"errors":[{"message":"access denied"}]}`
						} else {
							nodes := `[{"name":"zeta","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"shared","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"alpha","teamEnvironment":{"environment":{"name":"dev"}}}]`
							if tt.empty {
								nodes = `[]`
							}
							response = fmt.Sprintf(`{"data":{"team":{"openSearches":{"nodes":%s}}}}`, nodes)
						}
						if _, err := fmt.Fprint(w, response); err != nil {
							t.Errorf("writing response: %v", err)
						}
					}))
					defer server.Close()
					t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
					t.Setenv("HOME", t.TempDir())
					var output bytes.Buffer
					app, global, err := naistrix.NewApplication("test", "Test OpenSearch choices", "test", naistrix.ApplicationWithWriter(&output))
					if err != nil {
						t.Fatal(err)
					}
					global.Config = filepath.Join(t.TempDir(), "config.yaml")
					if tt.defaultEnvironment {
						if err := os.WriteFile(global.Config, []byte("environment: dev\n"), 0o600); err != nil {
							t.Fatal(err)
						}
					}
					additional := &flags.AdditionalFlags{}
					if err := app.AddGlobalFlags(additional); err != nil {
						t.Fatal(err)
					}
					root := OpenSearch(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
					var command *naistrix.Command
					for _, child := range root.SubCommands {
						if child.Name == action {
							command = child
						}
					}
					if command == nil {
						t.Fatal("command not found")
					}
					if command.AutoCompleteFunc != nil || (command.Args[0].ChoicesFunc != nil) != (action != "create") {
						t.Fatal("only existing-instance commands should use automatic choices")
					}
					runCalled := false
					// Exercise the real validators and choice lookup without mutating instances or requesting credentials.
					command.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
						runCalled = true
						if got := args.Get("name"); got != tt.value {
							t.Errorf("instance name = %q, want %q", got, tt.value)
						}
						return nil
					}
					if err := app.AddCommand(root); err != nil {
						t.Fatal(err)
					}
					args := []string{"opensearch", action}
					if !tt.omitTeam {
						args = append(args, "-t", "my-team")
					}
					if !tt.omitEnvironment && !tt.defaultEnvironment {
						args = append(args, "-e", "dev")
					}
					if action == "credentials" {
						args = append(args, "--permission", "READ", "--ttl", "1d")
					}
					if tt.completion {
						args = append([]string{"__complete"}, args...)
						if tt.afterValue {
							args = append(args, tt.value)
						}
						args = append(args, tt.prefix)
					} else {
						args = append(args, tt.value)
					}
					wantError := tt.wantError
					if action == "create" {
						wantError = ""
					}
					err = app.Run(naistrix.RunWithArgs(args))
					if tt.completion {
						if err != nil {
							t.Fatal(err)
						}
						got := output.String()
						if wantError != "" {
							if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, wantError) || strings.Count(got, "\n") != 2 {
								t.Errorf("completion = %q, want one active help error containing %q", got, wantError)
							}
						} else if tt.afterValue || action == "create" {
							if got != ":4\n" {
								t.Errorf("completion = %q, want no suggestions", got)
							}
						} else {
							wantNames := "alpha\nshared\nzeta\n"
							if tt.prefix != "" {
								wantNames = "shared\n"
							}
							want := wantNames + "_activeHelp_ " + command.Args[0].Prompt + "\n:4\n"
							if got != want {
								t.Errorf("completion = %q, want %q", got, want)
							}
						}
					} else if wantError != "" {
						if err == nil || !strings.Contains(err.Error(), wantError) {
							t.Errorf("error = %v, want %q", err, wantError)
						}
					} else if err != nil {
						t.Fatal(err)
					}
					if wantRun := !tt.completion && wantError == ""; runCalled != wantRun {
						t.Errorf("execution called = %v, want %v", runCalled, wantRun)
					}
					var wantOperations []string
					if action != "create" && !tt.omitTeam && !tt.omitEnvironment && !tt.afterValue {
						wantOperations = []string{"GetAllOpenSearches"}
					}
					if !slices.Equal(operations, wantOperations) {
						t.Errorf("API operations = %v, want %v", operations, wantOperations)
					}
				})
			}
		})
	}
}
