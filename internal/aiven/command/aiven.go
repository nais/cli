package command

import (
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

func Aiven(parentFlags *root.Flags) *cli.Command {
	aivenFlags := &flag.Aiven{Flags: parentFlags}
	return cli.NewCommand("aiven", "Manage Aiven services.", cli.WithSubCommands(
		create(aivenFlags),
		get(aivenFlags),
		tidy(aivenFlags),
	))
}
