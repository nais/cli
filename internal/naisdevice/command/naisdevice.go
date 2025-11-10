package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func Naisdevice(parentFlags *flags.GlobalFlags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "device",
		Title: "Interact with naisdevice",
		SubCommands: []*naistrix.Command{
			statuscmd(parentFlags),
			jitacmd(),
			doctorcmd(),
			disconnectcmd(),
			connectcmd(),
			configcmd(),
		},
	}
}
