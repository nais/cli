package commands

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

type aivenFlags struct{ *root.Flags }

func Command(parentFlags *root.Flags) *cli.Command {
	aivenFlags := &aivenFlags{Flags: parentFlags}
	return cli.NewCommand("aiven", "Manage Aiven services.", cli.WithSubCommands(
		create(aivenFlags),
		get(aivenFlags),
		tidy(aivenFlags),
	))
}
