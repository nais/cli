package command

import (
	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/root"
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
