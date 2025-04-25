package command

import (
	"github.com/urfave/cli/v3"
)

func Aiven() *cli.Command {
	return &cli.Command{
		Name:  "aiven",
		Usage: "Command used for management of AivenApplication",
		Commands: []*cli.Command{
			create(),
			get(),
			tidy(),
		},
	}
}
