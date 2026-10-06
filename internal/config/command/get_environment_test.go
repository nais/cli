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

func TestConfigGetEnvironment(t *testing.T) {
	for _, tt := range []struct {
		name, config, discoveryError, wantError string
		environments                            []string
		args                                    []string
		omitName, omitTeam                      bool
		wantOperations                          []string
	}{
		{name: "omitted environment resolves unique config", environments: []string{"dev-gcp"}, wantOperations: []string{"GetAllConfigs", "GetConfig"}},
		{name: "explicit environment bypasses discovery", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetConfig"}},
		{name: "default environment bypasses discovery", config: "environment: dev-gcp\n", wantOperations: []string{"GetConfig"}},
		{name: "explicit environment overrides default", config: "environment: prod-gcp\n", args: []string{"-e", "dev-gcp"}, wantOperations: []string{"GetConfig"}},
		{name: "missing config", wantError: `config "settings" not found`, wantOperations: []string{"GetAllConfigs"}},
		{name: "ambiguous config requires environment", environments: []string{"prod-gcp", "dev-gcp"}, wantError: "exists in multiple environments (dev-gcp, prod-gcp); specify -e, --environment", wantOperations: []string{"GetAllConfigs"}},
		{name: "discovery failure", discoveryError: "access denied", wantError: "access denied", wantOperations: []string{"GetAllConfigs"}},
		{name: "team remains required", omitTeam: true, wantError: "missing required team"},
		{name: "name remains required", omitName: true, wantError: "Expected exactly 1 argument"},
		{name: "key requires file", args: []string{"--key", "key"}, wantError: "--key is only used with --to-file"},
		{name: "file requires key", args: []string{"--to-file", "unused"}, wantError: "--to-file requires --key"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Name, EnvironmentName, TeamSlug string
						Filter                          struct{ Name string }
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetAllConfigs":
					if request.Variables.TeamSlug != "my-team" || request.Variables.Filter.Name != "settings" {
						t.Errorf("incorrect discovery target: %+v", request.Variables)
					}
					nodes := make([]string, 0, len(tt.environments))
					for _, env := range tt.environments {
						nodes = append(nodes, fmt.Sprintf(`{"name":"settings","teamEnvironment":{"environment":{"name":%q}}}`, env))
					}
					response = `{"data":{"team":{"configs":{"nodes":[` + strings.Join(nodes, ",") + `]}}}}`
					if tt.discoveryError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.discoveryError)
					}
				case "GetConfig":
					if request.Variables.Name != "settings" || request.Variables.TeamSlug != "my-team" || request.Variables.EnvironmentName != "dev-gcp" {
						t.Errorf("incorrect get target: %+v", request.Variables)
					}
					response = `{"data":{"team":{"environment":{"config":{"name":"settings","values":[{"name":"key","value":"value","encoding":"PLAINTEXT"}],"teamEnvironment":{"environment":{"name":"dev-gcp"}},"lastModifiedBy":{"email":"user@example.com"},"workloads":{"nodes":[]}}}}}}`
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
			app, global, err := naistrix.NewApplication("config-get-test", "Test config get", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Config(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "config", "get"}
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
				if detail.Name != "settings" || detail.Environment != "dev-gcp" || len(detail.Data) != 1 || detail.Data[0].Key != "key" || detail.Data[0].Value != "value" {
					t.Errorf("incorrect detail: %+v", detail)
				}
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
