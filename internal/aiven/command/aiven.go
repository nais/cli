package commands

import (
	"github.com/nais/cli/internal/aiven/command/create"
	"github.com/nais/cli/internal/aiven/flag"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

type aivenFlags struct{ *root.Flags }

func Aiven(parentFlags *root.Flags) *cli.Command {
	aivenFlags := &flag.Aiven{Flags: parentFlags}
	return cli.NewCommand("aiven", "Manage Aiven services.", cli.WithSubCommands(
		create.Create(aivenFlags),
		get(aivenFlags),
		tidy(aivenFlags),
	))
}
