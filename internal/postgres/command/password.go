package command

import (
	"github.com/urfave/cli/v2"
)

func password() *cli.Command {
	return &cli.Command{
		Name:  "password",
		Usage: "Administrate Postgres password",
		Subcommands: []*cli.Command{
			passwordRotate(),
		},
	}
}
