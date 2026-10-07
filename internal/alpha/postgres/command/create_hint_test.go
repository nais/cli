package command_test

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

	"github.com/nais/cli/internal/application"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestCreateHintKeepsAcceptedTarget(t *testing.T) {
	for _, tt := range []struct {
		name, config, envTeam, envEnvironment, wantScope string
		args                                             []string
		custom, relative, corruptAfterAcceptance         bool
	}{
		{name: "no defaults", config: "{}\n", args: []string{"-t", "my-team", "-e", "dev-gcp"}, wantScope: " -t my-team -e dev-gcp"},
		{name: "stored defaults match", config: "team: my-team\nenvironment: dev-gcp\n"},
		{name: "matching stored environment", config: "environment: dev-gcp\n", args: []string{"-t", "my-team"}, wantScope: " -t my-team"},
		{name: "explicit environment with only default team", config: "team: my-team\n", args: []string{"-e", "dev-gcp"}, wantScope: " -e dev-gcp"},
		{name: "explicit overrides mismatch defaults", config: "team: other-team\nenvironment: prod-gcp\n", args: []string{"-t", "my-team", "-e", "dev-gcp"}, wantScope: " -t my-team -e dev-gcp"},
		{name: "environment variable defaults match", config: "team: other-team\nenvironment: prod-gcp\n", envTeam: "my-team", envEnvironment: "dev-gcp"},
		{name: "explicit override of environment default", config: "{}\n", envTeam: "my-team", envEnvironment: "prod-gcp", args: []string{"-e", "dev-gcp"}, wantScope: " -e dev-gcp"},
		{name: "custom config preserved with shortened scope", custom: true, config: "team: my-team\nenvironment: dev-gcp\n"},
		{name: "relative custom config remains anchored after changing directory", custom: true, relative: true, config: "team: my-team\nenvironment: dev-gcp\n"},
		{name: "custom config preserved with explicit overrides", custom: true, config: "team: other-team\nenvironment: prod-gcp\n", args: []string{"-t", "my-team", "-e", "dev-gcp"}, wantScope: " -t my-team -e dev-gcp"},
		{name: "config read error cannot fail accepted creation", config: "team: my-team\nenvironment: dev-gcp\n", corruptAfterAcceptance: true, wantScope: " -t my-team -e dev-gcp"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("APPDATA", t.TempDir())
			t.Setenv("NAIS_TEAM", tt.envTeam)
			t.Setenv("NAIS_ENVIRONMENT", tt.envEnvironment)
			t.Setenv("NAIS_CONFIG", "")
			configDir, err := os.UserConfigDir()
			if err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(configDir, "nais", "config.yaml")
			if tt.custom {
				file = filepath.Join(t.TempDir(), "custom.yaml")
			}
			if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, []byte(tt.config), 0o600); err != nil {
				t.Fatal(err)
			}
			configArgument := file
			if tt.relative {
				t.Chdir(filepath.Dir(file))
				configArgument = filepath.Base(file)
			}
			var created gql.CreatePostgresInput
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Input                       gql.CreatePostgresInput `json:"input"`
						Team, Environment, Postgres string
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "CreatePostgres":
					created = request.Variables.Input
					if created.Name != "orders" || created.TeamSlug != "my-team" || created.EnvironmentName != "dev-gcp" {
						t.Errorf("wrong accepted target: %+v", created)
					}
					if tt.corruptAfterAcceptance {
						if err := os.WriteFile(file, []byte("team: [invalid"), 0o600); err != nil {
							t.Errorf("changing config after mutation: %v", err)
						}
					}
					response = `{"data":{"createPostgres":{"postgres":{"name":"created-orders"}}}}`
				case "GetPostgres":
					if request.Variables.Team != created.TeamSlug || request.Variables.Environment != created.EnvironmentName || request.Variables.Postgres != "created-orders" {
						t.Errorf("hint inspected different target: %+v; created %+v", request.Variables, created)
					}
					response = `{"data":{"team":{"environment":{"postgres":{"name":"created-orders","majorVersion":"18","highAvailability":false,"resources":{},"labels":[],"activeBranch":{"name":"main"},"desiredActiveBranch":"main","branches":{"nodes":[{"name":"main","state":"AVAILABLE"}]}}}}}}`
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
			var output bytes.Buffer
			app, _, err := application.New(&output)
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"alpha", "postgres", "create", "orders", "--yes", "--no-colors"}
			if tt.custom {
				args = append(args, "--config", configArgument)
			}
			args = append(args, tt.args...)
			if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
				t.Fatal(err)
			}
			want := "nais alpha postgres get created-orders"
			if tt.custom {
				want += " --config '" + file + "'"
			}
			want += tt.wantScope
			_, hint, found := strings.Cut(output.String(), "Check status: ")
			hint = strings.TrimSpace(hint)
			if !found || hint != want || !strings.Contains(output.String(), "created-orders · dev-gcp\nCreation requested.\n\n") {
				t.Fatalf("accepted output = %q; want hint %q", output.String(), want)
			}
			if strings.Contains(output.String(), "\x1b[") || strings.Contains(output.String(), "may still") {
				t.Errorf("unexpected styling or vague success language: %q", output.String())
			}
			if tt.corruptAfterAcceptance {
				if !slices.Equal(operations, []string{"CreatePostgres"}) {
					t.Errorf("API operations = %v", operations)
				}
				return
			}
			if tt.relative {
				otherDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(otherDir, filepath.Base(file)), []byte("team: wrong-team\nenvironment: wrong-env\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				t.Chdir(otherDir)
			}
			// The exact hint and quoting are checked above. Execute its equivalent arguments without a shell.
			hintArgs := strings.Fields(hint)[1:]
			if tt.custom {
				hintArgs = append([]string{"alpha", "postgres", "get", "created-orders", "--config", file}, strings.Fields(tt.wantScope)...)
			}
			var inspected bytes.Buffer
			freshApp, _, err := application.New(&inspected)
			if err != nil {
				t.Fatal(err)
			}
			if err := freshApp.Run(naistrix.RunWithArgs(hintArgs)); err != nil {
				t.Fatalf("emitted hint failed: %v", err)
			}
			if !slices.Equal(operations, []string{"CreatePostgres", "GetPostgres"}) {
				t.Errorf("API operations = %v", operations)
			}
		})
	}
}
