package command

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestRetiredBranchCommandsNeverCallAPI(t *testing.T) {
	for _, command := range []string{"create", "restore", "activate", "delete"} {
		t.Run(command, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				http.Error(w, "unexpected API request", http.StatusInternalServerError)
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("branch-retirement-test", "Test retired branches", "test", naistrix.ApplicationWithWriter(&output))
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
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "postgres", "branch", command, "orders", "main", "-t", "my-team", "-e", "dev-gcp"}
			if err := app.Run(naistrix.RunWithArgs(args)); err == nil || !strings.Contains(err.Error(), "unknown command") {
				t.Fatalf("retired branch command was not rejected as unknown: %v", err)
			}
			if calls.Load() != 0 {
				t.Fatal("retired branch command contacted the API")
			}
		})
	}
}
