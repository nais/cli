package command

import (
	"bytes"
	"encoding/json"
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

func TestSecretGetEnvironment(t *testing.T) {
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
	for _, tt := range []struct {
		name, config, wantError     string
		args                        []string
		omitName, omitTeam, managed bool
		wantOperations              []string
	}{
		{name: "environment is required before API calls", wantError: "missing required environment"},
		{name: "explicit environment", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetSecret"}},
		{name: "default environment", config: "environment: dev-gcp\n", wantOperations: []string{"GetSecret"}},
		{name: "explicit environment overrides default", config: "environment: prod-gcp\n", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetSecret"}},
		{name: "metadata does not expose values", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetSecret"}},
		{name: "platform managed secret uses explicit environment", args: []string{"-e", "dev-gcp", "--with-values", "--reason", "Investigating issue"}, managed: true, wantOperations: []string{"ViewSecretValues", "GetSecret"}},
		{name: "team remains required", omitTeam: true, wantError: "missing required team"},
		{name: "name remains required", args: []string{"-e", "dev-gcp"}, omitName: true, wantError: "Expected exactly 1 argument"},
		{name: "key requires file", args: []string{"-e", "dev-gcp", "--key", "key"}, wantError: "--key is only used with --to-file"},
		{name: "file requires key", args: []string{"-e", "dev-gcp", "--to-file", "unused"}, wantError: "--to-file requires --key"},
		{name: "reason requires values", args: []string{"-e", "dev-gcp", "--reason", "Investigating issue"}, wantError: "--reason can only be used"},
		{name: "short reason rejected", args: []string{"-e", "dev-gcp", "--with-values", "--reason", "short"}, wantError: "reason must be at least 10 characters"},
		{name: "values still require reason", args: []string{"-e", "dev-gcp", "--with-values"}, wantError: "prompting for reason"},
		{name: "environment required before auditing prompt", args: []string{"--with-values"}, wantError: "missing required environment"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Name, EnvironmentName, TeamSlug string
						Input                           struct{ Name, Environment, Team, Reason string }
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetSecret":
					if request.Variables.Name != "settings" || request.Variables.TeamSlug != "my-team" || request.Variables.EnvironmentName != "dev-gcp" {
						t.Errorf("incorrect get target: %+v", request.Variables)
					}
					response = `{"data":{"team":{"environment":{"secret":{"name":"settings","keys":["key"],"teamEnvironment":{"environment":{"name":"dev-gcp"}},"workloads":{"nodes":[]}}}}}}`
					if tt.managed {
						response = `{"errors":[{"message":"Resource not found"}]}`
					}
				case "ViewSecretValues":
					if request.Variables.Input.Name != "settings" || request.Variables.Input.Team != "my-team" || request.Variables.Input.Environment != "dev-gcp" || request.Variables.Input.Reason != "Investigating issue" {
						t.Errorf("incorrect audited value request: %+v", request.Variables.Input)
					}
					response = `{"data":{"viewSecretValues":{"values":[{"name":"key","value":"sensitive-value","encoding":"PLAINTEXT"}]}}}`
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
			t.Setenv("NAIS_TEAM", "")
			t.Setenv("NAIS_ENVIRONMENT", "")
			config := filepath.Join(t.TempDir(), "config.yaml")
			contents := tt.config
			if contents == "" {
				contents = "{}\n"
			}
			if err := os.WriteFile(config, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("secret-get-test", "Test secret get", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Secrets(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "secret", "get"}
			if !tt.omitName {
				args = append(args, "settings")
			}
			if !tt.omitTeam {
				args = append(args, "-t", "my-team")
			}
			args = append(args, "--output", "json")
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
				var detail struct {
					Name, Environment string
					Data              []Entry
				}
				if err := json.Unmarshal(output.Bytes(), &detail); err != nil {
					t.Fatalf("invalid JSON output %q: %v", output.String(), err)
				}
				if detail.Name != "settings" || detail.Environment != "dev-gcp" || len(detail.Data) != 1 || detail.Data[0].Key != "key" {
					t.Fatalf("incorrect detail: %+v", detail)
				}
				if !tt.managed && strings.Contains(output.String(), "sensitive-value") {
					t.Error("metadata-only output exposed a secret value")
				}
				if tt.managed && detail.Data[0].Value != "sensitive-value" {
					t.Error("missing platform-managed secret value")
				}
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}

func TestSecretGetCompletion(t *testing.T) {
	for _, tt := range []struct {
		name, response, want string
		args                 []string
		wantLookups          int
	}{
		{
			name:     "sorted names",
			response: `{"data":{"team":{"secrets":{"nodes":[{"name":"zeta","teamEnvironment":{"environment":{"name":"dev-gcp"}}},{"name":"settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}}]}}}}`,
			args:     []string{""}, want: "settings\nzeta\n_activeHelp_ Select a secret.\n:4\n", wantLookups: 1,
		},
		{
			name:     "prefix",
			response: `{"data":{"team":{"secrets":{"nodes":[{"name":"zeta","teamEnvironment":{"environment":{"name":"dev-gcp"}}},{"name":"settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}}]}}}}`,
			args:     []string{"set"}, want: "settings\n_activeHelp_ Select a secret.\n:4\n", wantLookups: 1,
		},
		{
			name:     "no matching prefix",
			response: `{"data":{"team":{"secrets":{"nodes":[{"name":"settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}}]}}}}`,
			args:     []string{"missing"}, want: "_activeHelp_ No secrets found.\n:4\n", wantLookups: 1,
		},
		{
			name:     "substring is not a prefix",
			response: `{"data":{"team":{"secrets":{"nodes":[{"name":"my-settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}},{"name":"settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}}]}}}}`,
			args:     []string{"set"}, want: "settings\n_activeHelp_ Select a secret.\n:4\n", wantLookups: 1,
		},
		{
			name:     "deduplicated and environment filtered",
			response: `{"data":{"team":{"secrets":{"nodes":[{"name":"settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}},{"name":"settings","teamEnvironment":{"environment":{"name":"dev-gcp"}}},{"name":"prod-only","teamEnvironment":{"environment":{"name":"prod-gcp"}}}]}}}}`,
			args:     []string{""}, want: "settings\n_activeHelp_ Select a secret.\n:4\n", wantLookups: 1,
		},
		{
			name: "empty list", response: `{"data":{"team":{"secrets":{"nodes":[]}}}}`,
			args: []string{""}, want: "_activeHelp_ No secrets found.\n:4\n", wantLookups: 1,
		},
		{
			name: "API error", response: `{"errors":[{"message":"access denied"}]}`,
			args: []string{""}, want: "Unable to fetch secrets for auto-completion:", wantLookups: 1,
		},
		{name: "after name", args: []string{"settings", ""}, want: ":4\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			lookups := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lookups++
				var request struct {
					OperationName string
					Variables     struct {
						TeamSlug string
						Filter   struct{ Environments []string }
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				if request.OperationName != "GetAllSecrets" || request.Variables.TeamSlug != "my-team" || !slices.Equal(request.Variables.Filter.Environments, []string{"dev-gcp"}) {
					t.Errorf("unexpected lookup: %+v", request)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(tt.response)); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test secret completion", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			root := Secrets(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
			command := root.SubCommands[slices.IndexFunc(root.SubCommands, func(c *naistrix.Command) bool { return c.Name == "get" })]
			if command.Args[0].ChoicesFunc != nil || command.AutoCompleteFunc == nil {
				t.Fatal("secret get must suggest names without restricting supplied names")
			}
			if err := app.AddCommand(root); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"__complete", "secret", "get", "-t", "my-team", "-e", "dev-gcp"}, tt.args...)
			if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != tt.want && (tt.name != "API error" || !strings.Contains(got, tt.want) || !strings.Contains(got, "access denied")) {
				t.Errorf("completion = %q, want %q", got, tt.want)
			}
			if lookups != tt.wantLookups {
				t.Errorf("lookups = %d, want %d", lookups, tt.wantLookups)
			}
		})
	}
}
