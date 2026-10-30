package command

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestPersonalAccessLevelCompletion(t *testing.T) {
	for _, command := range []string{"psql", "proxy"} {
		t.Run(command, func(t *testing.T) {
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("postgres-access-test", "Test personal access", "test", naistrix.ApplicationWithWriter(&output))
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
			if err := app.Run(naistrix.RunWithArgs([]string{"__complete", "postgres", command, "--access-level", ""})); err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(output.String(), "\n")
			for _, level := range []string{"read", "write", "admin", "maintain"} {
				if !slices.Contains(lines, level) {
					t.Errorf("completion missing %q: %s", level, output.String())
				}
			}
			if slices.Contains(lines, "dba") {
				t.Errorf("completion must not suggest dba: %s", output.String())
			}
		})
	}
}
