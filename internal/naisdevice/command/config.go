package command

import (
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func configcmd(rootFlags *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "config",
		Title: "Adjust or view the naisdevice configuration.",
		SubCommands: []*naistrix.Command{
			set(rootFlags),
			get(rootFlags),
		},
	}
}
