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

	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestPostgresCreateEnvironment(t *testing.T) {
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

	const hint = "missing required environment, specify an environment using `nais defaults set environment <environment>` or by using the -e, --environment flag"
	for _, tt := range []struct {
		name, config, response, createError, wantError string
		args                                           []string
		wantOperations                                 []string
	}{
		{name: "explicit environment bypasses discovery", args: []string{"-t", "my-team", "-e", "dev-gcp", "--yes"}, wantOperations: []string{"CreatePostgres"}},
		{name: "default environment bypasses discovery", config: "environment: dev-gcp\n", args: []string{"-t", "my-team", "--yes"}, wantOperations: []string{"CreatePostgres"}},
		{name: "explicit environment overrides default", config: "environment: prod-gcp\n", args: []string{"-t", "my-team", "-e", "dev-gcp", "--yes"}, wantOperations: []string{"CreatePostgres"}},
		{name: "multiple environments require a choice", args: []string{"-t", "my-team"}, response: `{"data":{"environments":{"nodes":[{"name":"prod-gcp"},{"name":"dev-gcp"}]}}}`, wantError: hint, wantOperations: []string{"Environments"}},
		{name: "yes does not choose from multiple environments", args: []string{"-t", "my-team", "--yes"}, response: `{"data":{"environments":{"nodes":[{"name":"prod-gcp"},{"name":"dev-gcp"}]}}}`, wantError: hint, wantOperations: []string{"Environments"}},
		{name: "one environment still requires a choice", args: []string{"-t", "my-team"}, response: `{"data":{"environments":{"nodes":[{"name":"dev-gcp"}]}}}`, wantError: hint, wantOperations: []string{"Environments"}},
		{name: "yes does not choose the sole environment", args: []string{"-t", "my-team", "--yes"}, response: `{"data":{"environments":{"nodes":[{"name":"dev-gcp"}]}}}`, wantError: hint, wantOperations: []string{"Environments"}},
		{name: "no environments", args: []string{"-t", "my-team", "--yes"}, response: `{"data":{"environments":{"nodes":[]}}}`, wantError: hint, wantOperations: []string{"Environments"}},
		{name: "discovery failure is propagated", args: []string{"-t", "my-team", "--yes"}, response: `{"errors":[{"message":"access denied"}]}`, wantError: "fetching environments:", wantOperations: []string{"Environments"}},
		{name: "team is required before discovery", args: []string{"--yes"}, wantError: "missing required team"},
		{name: "invalid version fails before discovery", args: []string{"-t", "my-team", "--version", "17", "--yes"}, wantError: "--version must be 18"},
		{name: "confirmation unavailable stops creation", args: []string{"-t", "my-team", "-e", "dev-gcp"}, wantError: "no interactive terminal available"},
		{name: "rejected mutation has no hint", args: []string{"-t", "my-team", "-e", "dev-gcp", "--yes"}, createError: "creation rejected", wantError: "creation rejected", wantOperations: []string{"CreatePostgres"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			var created gql.CreatePostgresInput
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Input gql.CreatePostgresInput `json:"input"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				response := tt.response
				switch request.OperationName {
				case "Environments":
				case "CreatePostgres":
					created = request.Variables.Input
					response = `{"data":{"createPostgres":{"postgres":{"name":"orders"}}}}`
					if tt.createError != "" {
						response = `{"errors":[{"message":"creation rejected"}]}`
					}
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
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("APPDATA", t.TempDir())
			t.Setenv("NAIS_TEAM", "")
			t.Setenv("NAIS_ENVIRONMENT", "")
			t.Setenv("NAIS_CONFIG", "")
			config := filepath.Join(t.TempDir(), "config.yaml")
			contents := tt.config
			if contents == "" {
				contents = "{}\n"
			}
			if err := os.WriteFile(config, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("nais", "Test Postgres creation", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			parent := &flag.Postgres{GlobalFlags: &flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional}}
			if err := app.AddCommand(instanceCreateCommand(parent)); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--config", config, "--no-colors", "create", "orders"}, tt.args...)
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v; want %q", err, tt.wantError)
				}
				if strings.Contains(output.String(), "Check status:") || strings.Contains(output.String(), "Creation requested.") {
					t.Errorf("accepted output shown for failure: %s", output.String())
				}
				if tt.name == "discovery failure is propagated" && (err == nil || !strings.Contains(err.Error(), "access denied")) {
					t.Errorf("missing API error: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if created.Name != "orders" || created.TeamSlug != "my-team" || created.EnvironmentName != "dev-gcp" || created.MajorVersion != "18" {
					t.Errorf("unexpected create input: %+v", created)
				}
				wantHint := "Check status: nais alpha postgres get orders --config " + quoteShellArgument(config) + " -t my-team"
				if tt.name != "default environment bypasses discovery" {
					wantHint += " -e dev-gcp"
				}
				if !strings.Contains(output.String(), "Environment") || !strings.Contains(output.String(), "dev-gcp") || strings.Contains(output.String(), "prod-gcp") || !strings.Contains(output.String(), "orders · dev-gcp\nCreation requested.\n\n") || !strings.Contains(output.String(), wantHint+"\n") {
					t.Errorf("missing or incorrect destination in confirmation/result: %s", output.String())
				}
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
