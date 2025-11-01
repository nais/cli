package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/naistrix"
)

func Naisdevice(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Device{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:  "device",
		Title: "Interact with naisdevice",
		SubCommands: []*naistrix.Command{
			statuscmd(flags),
			gatewaycmd(flags),
			jitacmd(),
			doctorcmd(),
			disconnectcmd(),
			connectcmd(),
			configcmd(),
		},
	}
}
