package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestApplicationIssuesEmptyOutput(t *testing.T) {
	const noMatchingIssues = `[{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}},"issues":{"nodes":[]}}]`
	for _, tt := range []struct {
		name, format, applications, want string
	}{
		{name: "no matching issues in JSON", format: "json", applications: noMatchingIssues, want: "[]"},
		{name: "no matching issues in table", format: "table", applications: noMatchingIssues, want: "No issues found for application."},
		{name: "no application nodes in JSON", format: "json", applications: `[]`, want: "[]"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string
					Variables     struct {
						Filter struct{ IssueType gql.IssueType }
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				var nodes string
				switch request.OperationName {
				case "GetApplicationNames":
					nodes = `[{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}}}]`
				case "GetApplicationIssues":
					if got := request.Variables.Filter.IssueType; got != gql.IssueTypeApplicationRestartLoop {
						t.Errorf("issue type = %q, want %q", got, gql.IssueTypeApplicationRestartLoop)
					}
					nodes = tt.applications
				default:
					t.Errorf("unexpected API operation: %s", request.OperationName)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprintf(w, `{"data":{"team":{"applications":{"nodes":%s}}}}`, nodes); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

			var stdout bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test application issues", "test", naistrix.ApplicationWithWriter(&stdout))
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
			args := []string{"app", "issues", "orders", "-t", "my-team", "-e", "dev", "--issuetype", "application_restart_loop", "--output", tt.format}
			if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(stdout.String()); got != tt.want {
				t.Errorf("stdout = %q, want %q", got, tt.want)
			}
		})
	}
}
