package command

import (
	"github.com/urfave/cli/v2"
)

func Device() *cli.Command {
	return &cli.Command{
		Name:  "device",
		Usage: "Command used for management of naisdevice",
		Subcommands: []*cli.Command{
			config(),
			connect(),
			disconnect(),
			jita(),
			status(),
			doctor(),
		},
	}
}
