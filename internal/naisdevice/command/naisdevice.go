package command

import (
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
)

func Naisdevice(rootFlags *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "device",
		Title: "Interact with naisdevice",
		SubCommands: []*cli.Command{
			statuscmd(rootFlags),
			jitacmd(rootFlags),
			doctorcmd(rootFlags),
			disconnectcmd(rootFlags),
			connectcmd(rootFlags),
			configcmd(rootFlags),
		},
	}
}
