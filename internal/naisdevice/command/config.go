package command

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

func configcmd(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("config", "Adjust or view the naisdevice configuration.", cli.WithSubCommands(
		set(rootFlags),
		get(rootFlags),
	))
}
