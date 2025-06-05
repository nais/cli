package command

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

func configcmd(rootFlags *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Short: "Adjust or view the naisdevice configuration.",
		SubCommands: []*cli.Command{
			set(rootFlags),
			get(rootFlags),
		},
	}
}
