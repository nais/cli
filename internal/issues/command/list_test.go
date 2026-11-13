package command

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/issues"
	"github.com/nais/naistrix"
)

func TestPostgresIssueListOutput(t *testing.T) {
	for _, format := range []string{"table", "json"} {
		t.Run(format, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"team":{"issues":{"nodes":[{"__typename":"PostgresIssue","id":"issue-orders","severity":"CRITICAL","message":"Database failure","teamEnvironment":{"environment":{"name":"dev"}},"postgres":{"__typename":"Postgres","name":"orders"}}]}}}}`))
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("NAIS_API_LOCAL_EMAIL", "probe@example.com")
			t.Setenv("HOME", t.TempDir())
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("issues-test", "Test issue resources", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Issues(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(naistrix.RunWithArgs([]string{"--config", config, "--no-colors", "issues", "list", "-t", "team", "--output", format, "--issuetype", "POSTGRES", "--resourcetype", "POSTGRES"})); err != nil {
				t.Fatal(err)
			}
			if format == "json" {
				var rows []issues.Issue
				if err := json.Unmarshal(output.Bytes(), &rows); err != nil {
					t.Fatal(err)
				}
				if len(rows) != 1 || rows[0].ResourceName != "orders" || rows[0].ResourceType != "Postgres" || rows[0].Environment != "dev" {
					t.Errorf("unexpected issue resources: %s", output.String())
				}
			} else {
				for _, want := range []string{"orders", "Postgres", "dev", "Database failure"} {
					if !strings.Contains(output.String(), want) {
						t.Errorf("table does not identify %q: %s", want, output.String())
					}
				}
			}
			if strings.Contains(strings.ToLower(output.String()), "branch") {
				t.Errorf("issues must identify the database, not a branch: %s", output.String())
			}
		})
	}
}

func TestPostgresIssueResourceURL(t *testing.T) {
	got := issueResourceURL("console.example.com", "team", "dev", "Postgres", "orders")
	want := "https://console.example.com/team/team/dev/postgres/orders"
	if got != want {
		t.Errorf("database issue link = %q, want %q", got, want)
	}
}
