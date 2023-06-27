package deviceCmd

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "device",
		Aliases:         []string{"d"},
		Description:     "Command used for management of naisdevice",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			configCommand(),
			connectCommand(),
			disconnectCommand(),
			jitaCommand(),
			statusCommand(),
		},
	}
}
