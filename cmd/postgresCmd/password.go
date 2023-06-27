package postgresCmd

import (
	"github.com/urfave/cli/v2"
)

func passwordCommand() *cli.Command {
	return &cli.Command{
		Name:  "password",
		Usage: "Administrate Postgres password",
		Subcommands: []*cli.Command{
			passwordRotateCommand(),
		},
	}
}
