package command

import (
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
)

func configcmd(rootFlags *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Title: "Adjust or view the naisdevice configuration.",
		SubCommands: []*cli.Command{
			set(rootFlags),
			get(rootFlags),
		},
	}
}
