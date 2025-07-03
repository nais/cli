package command

import (
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Naisdevice(rootFlags *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "device",
		Title: "Interact with naisdevice",
		SubCommands: []*naistrix.Command{
			statuscmd(rootFlags),
			jitacmd(rootFlags),
			doctorcmd(rootFlags),
			disconnectcmd(rootFlags),
			connectcmd(rootFlags),
			configcmd(rootFlags),
		},
	}
}
