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
		name, config, discoveryError, wantError string
		environments                            []string
		args                                    []string
		omitName, omitTeam, managed             bool
		wantOperations                          []string
	}{
		{name: "omitted environment resolves unique secret", environments: []string{"dev-gcp"}, wantOperations: []string{"GetAllSecrets", "GetSecret"}},
		{name: "explicit environment bypasses discovery", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetSecret"}},
		{name: "default environment bypasses discovery", config: "environment: dev-gcp\n", wantOperations: []string{"GetSecret"}},
		{name: "explicit environment overrides default", config: "environment: prod-gcp\n", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetSecret"}},
		{name: "missing secret", wantError: `secret "settings" not found`, wantOperations: []string{"GetAllSecrets"}},
		{name: "ambiguous secret requires environment", environments: []string{"prod-gcp", "dev-gcp"}, wantError: "exists in multiple environments (dev-gcp, prod-gcp); specify -e, --environment", wantOperations: []string{"GetAllSecrets"}},
		{name: "discovery failure", discoveryError: "access denied", wantError: "access denied", wantOperations: []string{"GetAllSecrets"}},
		{name: "metadata does not expose values", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetSecret"}},
		{name: "platform managed secret uses explicit environment", args: []string{"-e", "dev-gcp", "--with-values", "--reason", "Investigating issue"}, managed: true, wantOperations: []string{"ViewSecretValues", "GetSecret"}},
		{name: "team remains required", omitTeam: true, wantError: "missing required team"},
		{name: "name remains required", omitName: true, wantError: "Expected exactly 1 argument"},
		{name: "key requires file", args: []string{"--key", "key"}, wantError: "--key is only used with --to-file"},
		{name: "file requires key", args: []string{"--to-file", "unused"}, wantError: "--to-file requires --key"},
		{name: "reason requires values", args: []string{"--reason", "Investigating issue"}, wantError: "--reason can only be used"},
		{name: "short reason rejected", args: []string{"--with-values", "--reason", "short"}, wantError: "reason must be at least 10 characters"},
		{name: "values still require reason", environments: []string{"dev-gcp"}, args: []string{"--with-values"}, wantError: "prompting for reason", wantOperations: []string{"GetAllSecrets"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Name, EnvironmentName, TeamSlug string
						Filter                          struct{ Name string }
						Input                           struct{ Name, Environment, Team, Reason string }
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetAllSecrets":
					if request.Variables.TeamSlug != "my-team" || request.Variables.Filter.Name != "settings" {
						t.Errorf("incorrect discovery target: %+v", request.Variables)
					}
					nodes := make([]string, 0, len(tt.environments))
					for _, env := range tt.environments {
						nodes = append(nodes, fmt.Sprintf(`{"name":"settings","teamEnvironment":{"environment":{"name":%q}}}`, env))
					}
					response = `{"data":{"team":{"secrets":{"nodes":[` + strings.Join(nodes, ",") + `]}}}}`
					if tt.discoveryError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.discoveryError)
					}
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
