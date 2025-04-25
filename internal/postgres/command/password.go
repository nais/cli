package command

import (
	"github.com/urfave/cli/v3"
)

func password() *cli.Command {
	return &cli.Command{
		Name:  "password",
		Usage: "Administrate Postgres password",
		Commands: []*cli.Command{
			passwordRotate(),
		},
	}
}
