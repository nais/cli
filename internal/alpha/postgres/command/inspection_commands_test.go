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
		name     string
		args     []string
		rejected bool
	}{
		{name: "help lists get instead of top-level status", args: []string{"alpha", "postgres", "--help"}},
		{name: "completion lists get instead of top-level status", args: []string{"__complete", "alpha", "postgres", ""}},
		{name: "removed command is not aliased", args: []string{"alpha", "postgres", "status", "orders"}, rejected: true},
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
			if tt.rejected {
				if err == nil {
					t.Error("removed top-level status command was accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var get, branch bool
			for line := range strings.SplitSeq(output.String(), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				get = get || fields[0] == "get"
				branch = branch || fields[0] == "branch"
				if fields[0] == "status" {
					t.Errorf("removed command is visible: %s", output.String())
				}
			}
			if !get || !branch {
				t.Errorf("missing get/branch commands: %s", output.String())
			}
		})
	}
}
