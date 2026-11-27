package command

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestApplicationEventsNameLookup(t *testing.T) {
	for _, tt := range []struct {
		name          string
		args          []string
		wantOperation string
		wantOutput    string
	}{
		{
			name: "explicit name uses scoped query", args: []string{"app", "events", "older-app", "-e", "dev"},
			wantOperation: "GetApplicationEvents", wantOutput: "No retained events in this range. The API retains workload events for 30 days.\n",
		},
		{
			name: "completion without environment deduplicates search results", args: []string{"__complete", "app", "events", "ord"},
			wantOperation: "SearchWorkloadNames", wantOutput: "orders\norders-canary\n:4\n",
		},
		{
			name: "completion with environment filters results", args: []string{"__complete", "app", "events", "-e", "dev", "ord"},
			wantOperation: "SearchWorkloadNames", wantOutput: "orders\n:4\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string
					Query         string
					Variables     struct {
						Team, Env, Name string
						Filter          gql.SearchFilter
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "SearchWorkloadNames":
					f := request.Variables.Filter
					if !strings.Contains(request.Query, "search(first: 30, filter: $filter)") || f.Query != "ord" || !slices.Equal(f.Types, []gql.SearchType{gql.SearchTypeApplication}) || f.TeamScope == nil || !slices.Equal(f.TeamScope.TeamSlugs, []string{"my-team"}) {
						t.Errorf("unexpected completion query: %s (%+v)", request.Query, f)
					}
					response = `{"data":{"search":{"nodes":[
						{"__typename":"Application","name":"orders","teamEnvironment":{"environment":{"name":"prod"}}},
						{"__typename":"Application","name":"orders-canary","teamEnvironment":{"environment":{"name":"prod"}}},
						{"__typename":"Application","name":"orders","teamEnvironment":{"environment":{"name":"dev"}}}
					]}}}`
				case "GetApplicationEvents":
					if got := request.Variables; got.Team != "my-team" || got.Env != "dev" || got.Name != "older-app" {
						t.Errorf("events query target = %+v, want my-team/dev/older-app", got)
					}
					response = `{"data":{"team":{"environment":{"application":{"events":{"pageInfo":{"hasNextPage":false},"nodes":[]}}}}}}`
				default:
					t.Errorf("unexpected API operation: %s", request.OperationName)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := io.WriteString(w, response); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

			var stdout bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test application events", "test", naistrix.ApplicationWithWriter(&stdout))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(App(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(naistrix.RunWithArgs(append([]string{"-t", "my-team"}, tt.args...))); err != nil {
				t.Fatal(err)
			}
			server.Close()
			if got := stdout.String(); got != tt.wantOutput {
				t.Errorf("stdout = %q, want %q", got, tt.wantOutput)
			}
			if !slices.Equal(operations, []string{tt.wantOperation}) {
				t.Errorf("API operations = %v, want [%s]", operations, tt.wantOperation)
			}
		})
	}
}
