package command

import (
	"errors"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func TestPostgresCommandsRequireTeam(t *testing.T) {
	for _, tt := range []struct {
		name    string
		command func(*flag.Postgres) *naistrix.Command
	}{
		{name: "get", command: instanceGetCommand},
		{name: "create", command: instanceCreateCommand},
		{name: "update", command: instanceUpdateCommand},
		{name: "delete", command: instanceDeleteCommand},
		{name: "branch list", command: branchListCommand},
		{name: "branch create", command: branchCreateCommand},
		{name: "branch activate", command: branchActivateCommand},
		{name: "branch delete", command: branchDeleteCommand},
		{name: "psql", command: psqlCommand},
		{name: "proxy", command: proxyCommand},
	} {
		t.Run(tt.name, func(t *testing.T) {
			additional := &flags.AdditionalFlags{}
			cmd := tt.command(&flag.Postgres{GlobalFlags: &flags.GlobalFlags{AdditionalFlags: additional}})
			if cmd.ValidateFlagsFunc == nil {
				t.Fatal("command must validate flags before resolving arguments or running")
			}
			if err := cmd.ValidateFlagsFunc(t.Context()); !errors.Is(err, validation.ErrMissingTeam) {
				t.Errorf("error = %v, want ErrMissingTeam", err)
			}
			additional.Team = "my-team"
			if err := cmd.ValidateFlagsFunc(t.Context()); err != nil {
				t.Errorf("validation with team and no explicit environment failed: %v", err)
			}
		})
	}
}
