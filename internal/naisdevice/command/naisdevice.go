package command

import (
	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/root"
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
