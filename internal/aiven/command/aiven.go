package command

import (
	"github.com/urfave/cli/v2"
)

func Aiven() *cli.Command {
	return &cli.Command{
		Name:  "aiven",
		Usage: "Command used for management of AivenApplication",
		Subcommands: []*cli.Command{
			create(),
			get(),
			tidy(),
		},
	}
}
