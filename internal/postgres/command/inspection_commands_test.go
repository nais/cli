package command_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/application"
	"github.com/nais/naistrix"
)

func TestPostgresInspectionHelpAndCompletion(t *testing.T) {
	for _, tt := range []struct {
		name      string
		args      []string
		wantError string
		wantGet   bool
	}{
		{name: "help lists get instead of top-level status", args: []string{"postgres", "--help"}, wantGet: true},
		{name: "completion lists get instead of top-level status", args: []string{"__complete", "postgres", ""}, wantGet: true},
		{name: "removed command is not aliased", args: []string{"postgres", "status", "orders"}, wantError: `unknown command "status"`},
		{name: "branch command is removed", args: []string{"postgres", "branch", "list", "orders"}, wantError: `unknown command "branch"`},
		{name: "psql branch flag is removed", args: []string{"postgres", "psql", "orders", "--branch", "main"}, wantError: "unknown flag: --branch"},
		{name: "proxy branch flag is removed", args: []string{"postgres", "proxy", "orders", "--branch", "main"}, wantError: "unknown flag: --branch"},
		{name: "psql help hides branches", args: []string{"postgres", "psql", "--help"}},
		{name: "proxy help hides branches", args: []string{"postgres", "proxy", "--help"}},
		{name: "psql flag completion hides branches", args: []string{"__complete", "postgres", "psql", "--"}},
		{name: "proxy flag completion hides branches", args: []string{"__complete", "postgres", "proxy", "--"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("APPDATA", t.TempDir())
			t.Setenv("NAIS_CONFIG", "")
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, _, err := application.New(&output)
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--config", config, "--no-colors"}, tt.args...)
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var get bool
			for line := range strings.SplitSeq(output.String(), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				get = get || fields[0] == "get"
				if fields[0] == "status" || fields[0] == "branch" || strings.Contains(line, "--branch") {
					t.Errorf("removed command is visible: %s", output.String())
				}
			}
			if tt.wantGet && !get {
				t.Errorf("missing get command: %s", output.String())
			}
		})
	}
}
