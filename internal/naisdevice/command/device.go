package command

import (
	"github.com/urfave/cli/v3"
)

func Device() *cli.Command {
	return &cli.Command{
		Name:  "device",
		Usage: "Command used for management of naisdevice",
		Commands: []*cli.Command{
			config(),
			connect(),
			disconnect(),
			jita(),
			status(),
			doctor(),
		},
	}
}
