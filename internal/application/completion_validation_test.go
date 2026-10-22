package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nais/naistrix"
)

func TestPositionalCompletionValidatesFlagsBeforeAPI(t *testing.T) {
	for _, tt := range []struct {
		name, wantError string
		args            []string
	}{
		{name: "app team", args: []string{"app", "status"}, wantError: "missing required team"},
		{name: "job team", args: []string{"job", "trigger", "-e", "dev"}, wantError: "missing required team"},
		{name: "job environment", args: []string{"job", "trigger", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "job run environment", args: []string{"job", "run", "delete", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "config team", args: []string{"config", "get"}, wantError: "missing required team"},
		{name: "config environment", args: []string{"config", "delete", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "secret team", args: []string{"secret", "get"}, wantError: "missing required team"},
		{name: "secret environment", args: []string{"secret", "delete", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "Kafka team", args: []string{"kafka", "list-grants", "-e", "dev"}, wantError: "missing required team"},
		{name: "Kafka environment", args: []string{"kafka", "list-grants", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "Kafka revoke environment", args: []string{"kafka", "revoke-grant", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "OpenSearch team", args: []string{"opensearch", "get", "-e", "dev"}, wantError: "missing required team"},
		{name: "OpenSearch environment", args: []string{"opensearch", "get", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "Valkey team", args: []string{"valkey", "get", "-e", "dev"}, wantError: "missing required team"},
		{name: "Valkey environment", args: []string{"valkey", "get", "-t", "my-team"}, wantError: "missing required environment"},
		{name: "Postgres team", args: []string{"postgres", "update"}, wantError: "missing required team"},
		{name: "Postgres branch team", args: []string{"postgres", "branch", "activate", "orders"}, wantError: "missing required team"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				http.Error(w, "unexpected API request", http.StatusInternalServerError)
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			var output bytes.Buffer
			app, flags, err := New(&output)
			if err != nil {
				t.Fatal(err)
			}
			flags.Config = filepath.Join(t.TempDir(), "config.yaml")
			args := append([]string{"__complete"}, tt.args...)
			args = append(args, "")
			if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
				t.Fatal(err)
			}
			got := output.String()
			if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, tt.wantError) || !strings.HasSuffix(got, "\n:4\n") {
				t.Errorf("completion = %q, want active help containing %q and NoFileComp", got, tt.wantError)
			}
			if calls.Load() != 0 {
				t.Errorf("API calls = %d, want none before flag validation succeeds", calls.Load())
			}
		})
	}
}

func TestPositionalCompletionEnvironmentFiltering(t *testing.T) {
	for _, tt := range []struct {
		name, operation, field string
		args                   []string
	}{
		{name: "config", operation: "GetAllConfigs", field: "configs", args: []string{"config", "get"}},
		{name: "secret", operation: "GetAllSecrets", field: "secrets", args: []string{"secret", "get"}},
		{name: "job", operation: "GetJobNames", field: "jobs", args: []string{"job", "activity"}},
	} {
		for _, environment := range []string{"", "dev"} {
			t.Run(tt.name+"/environment="+environment, func(t *testing.T) {
				var operations []string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var request struct {
						OperationName string `json:"operationName"`
						Variables     struct {
							Team, TeamSlug string
							Filter         struct{ Environments []string }
						} `json:"variables"`
					}
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Errorf("decoding request: %v", err)
					}
					operations = append(operations, request.OperationName)
					if request.Variables.Team != "my-team" && request.Variables.TeamSlug != "my-team" {
						t.Errorf("incorrect team in completion request: %+v", request.Variables)
					}
					if tt.name != "job" {
						var wantEnvironments []string
						if environment != "" {
							wantEnvironments = []string{environment}
						}
						if !slices.Equal(request.Variables.Filter.Environments, wantEnvironments) {
							t.Errorf("environment filter = %v, want %v", request.Variables.Filter.Environments, wantEnvironments)
						}
					}
					w.Header().Set("Content-Type", "application/json")
					nodes := `[{"name":"shared","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"shared","teamEnvironment":{"environment":{"name":"prod"}}},{"name":"prod-only","teamEnvironment":{"environment":{"name":"prod"}}}]`
					if tt.name != "job" {
						// Config and secret environment filtering is performed by the API.
						nodes = `[{"name":"shared","teamEnvironment":{"environment":{"name":"dev"}}}]`
					}
					_, err := fmt.Fprintf(w, `{"data":{"team":{%q:{"nodes":%s}}}}`, tt.field, nodes)
					if err != nil {
						t.Errorf("writing response: %v", err)
					}
				}))
				defer server.Close()
				t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
				t.Setenv("HOME", t.TempDir())
				var output bytes.Buffer
				app, flags, err := New(&output)
				if err != nil {
					t.Fatal(err)
				}
				flags.Config = filepath.Join(t.TempDir(), "config.yaml")
				args := append([]string{"__complete"}, tt.args...)
				args = append(args, "-t", "my-team")
				if environment != "" {
					args = append(args, "-e", environment)
				}
				args = append(args, "")
				if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
					t.Fatal(err)
				}
				got := output.String()
				if environment == "" && tt.name != "job" {
					if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, "missing required environment") || !strings.HasSuffix(got, "\n:4\n") || strings.Count(got, "\n") != 2 {
						t.Errorf("completion = %q, want only environment guidance and NoFileComp", got)
					}
					if len(operations) != 0 {
						t.Errorf("API operations = %v, want none without an environment", operations)
					}
					return
				}
				if strings.Count(got, "shared\n") != 1 {
					t.Errorf("completion = %q, want one shared-name suggestion", got)
				}
				if environment == "" {
					want := "prod-only\nshared\n_activeHelp_ Name of the job to show activity for\n:4\n"
					if got != want {
						t.Errorf("completion = %q, want %q", got, want)
					}
				} else if strings.Contains(got, "prod-only\n") {
					t.Errorf("completion = %q, unexpected environment filtering", got)
				}
				if !slices.Equal(operations, []string{tt.operation}) {
					t.Errorf("API operations = %v, want [%s]", operations, tt.operation)
				}
			})
		}
	}
}
