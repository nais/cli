package command

import (
	"github.com/nais/naistrix"
)

func configcmd() *naistrix.Command {
	return &naistrix.Command{
		Name:        "config",
		Title:       "Adjust or view the naisdevice configuration.",
		Description: "Commands for getting and setting naisdevice configuration values such as autoconnect.",
		SubCommands: []*naistrix.Command{
			set(),
			get(),
		},
	}
}
