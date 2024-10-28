package aivencmd

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "aiven",
		Usage: "Command used for management of AivenApplication",
		Subcommands: []*cli.Command{
			createCommand(),
			getCommand(),
			tidyCommand(),
		},
	}
}
