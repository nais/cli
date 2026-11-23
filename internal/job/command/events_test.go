package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestJobEventsNameLookup(t *testing.T) {
	for _, tt := range []struct {
		name          string
		args          []string
		wantOperation string
		wantOutput    string
		wantError     string
	}{
		{
			name: "explicit job not in completion page", args: []string{"job", "events", "older-job"},
			wantOperation: "GetJobEvents", wantOutput: "No retained events in this range. The API retains workload events for 30 days.\n",
		},
		{
			name: "missing job is reported by the events query", args: []string{"job", "events", "missing-job"},
			wantOperation: "GetJobEvents", wantError: "job not found",
		},
		{
			name: "job name completion", args: []string{"__complete", "job", "events", "ba"},
			wantOperation: "GetJobNames", wantOutput: "batch\n_activeHelp_ Select a job.\n:4\n",
		},
		{
			name: "no completion after name", args: []string{"__complete", "job", "events", "older-job", ""},
			wantOutput: ":4\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string
					Variables     struct{ Team, Env, Name string }
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				operations = append(operations, request.OperationName)
				if request.Variables.Team != "my-team" {
					t.Errorf("team = %q, want my-team", request.Variables.Team)
				}
				var response string
				switch request.OperationName {
				case "GetJobNames":
					response = `{"data":{"team":{"jobs":{"nodes":[{"name":"batch","teamEnvironment":{"environment":{"name":"dev"}}},{"name":"backup","teamEnvironment":{"environment":{"name":"prod"}}},{"name":"cleanup","teamEnvironment":{"environment":{"name":"dev"}}}]}}}}`
				case "GetJobEvents":
					if got := request.Variables; got.Env != "dev" || got.Name != tt.args[2] {
						t.Errorf("events query target = %+v, want dev/%s", got, tt.args[2])
					}
					if tt.wantError != "" {
						response = `{"errors":[{"message":"job not found"}]}`
					} else {
						response = `{"data":{"team":{"environment":{"job":{"events":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}}`
					}
				default:
					t.Errorf("unexpected API operation: %s", request.OperationName)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprint(w, response); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

			var stdout bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test job events", "test", naistrix.ApplicationWithWriter(&stdout))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Job(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"-t", "my-team", "-e", "dev"}, tt.args...)
			err = app.Run(naistrix.RunWithArgs(args))
			server.Close()
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			} else if got := stdout.String(); got != tt.wantOutput {
				t.Errorf("stdout = %q, want %q", got, tt.wantOutput)
			}
			var wantOperations []string
			if tt.wantOperation != "" {
				wantOperations = []string{tt.wantOperation}
			}
			if !slices.Equal(operations, wantOperations) {
				t.Errorf("API operations = %v, want %v", operations, wantOperations)
			}
		})
	}
}
