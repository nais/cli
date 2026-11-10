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
	"github.com/nais/naistrix"
)

func TestPostgresListHidesBranches(t *testing.T) {
	for _, format := range []string{"table", "json"} {
		t.Run(format, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct{ OperationName, Query string }
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
				}
				if request.OperationName != "GetTeamPostgreses" || strings.Contains(strings.ToLower(request.Query), "branch") {
					t.Errorf("unexpected list query: %+v", request)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}},"majorVersion":"18","highAvailability":true,"activeBranch":{"name":"hidden-branch"}}],"pageInfo":{"hasNextPage":false}}}}}`)); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("postgres-list-test", "Test Postgres listing", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Postgres(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(naistrix.RunWithArgs([]string{"--config", config, "--no-colors", "postgres", "list", "-t", "my-team", "--output", format})); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(strings.ToLower(output.String()), "branch") {
				t.Errorf("list exposes branches: %s", output.String())
			}
			for _, want := range []string{"orders", "dev-gcp", "18"} {
				if !strings.Contains(output.String(), want) {
					t.Errorf("list missing %q: %s", want, output.String())
				}
			}
			if format == "json" {
				var instances []map[string]json.RawMessage
				if err := json.Unmarshal(output.Bytes(), &instances); err != nil {
					t.Fatal(err)
				}
				if len(instances) != 1 || len(instances[0]) != 4 {
					t.Errorf("unexpected instance fields: %s", output.String())
				}
			}
		})
	}
}
