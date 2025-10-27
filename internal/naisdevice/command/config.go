package command

import (
	"github.com/nais/naistrix"
)

func configcmd() *naistrix.Command {
	return &naistrix.Command{
		Name:  "config",
		Title: "Adjust or view the naisdevice configuration.",
		SubCommands: []*naistrix.Command{
			set(),
			get(),
		},
	}
}
