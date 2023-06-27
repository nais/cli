package aivenCmd

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "aiven",
		Aliases:     []string{"a"},
		Description: "Command used for management of AivenApplication",
		Subcommands: []*cli.Command{
			createCommand(),
			getCommand(),
			tidyCommand(),
		},
	}
}
