package devicecmd

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "device",
		Usage: "Command used for management of naisdevice",
		Subcommands: []*cli.Command{
			configCommand(),
			connectCommand(),
			disconnectCommand(),
			jitaCommand(),
			statusCommand(),
		},
	}
}
