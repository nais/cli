package postgresCmd

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "postgres",
		Usage: "Command used for connecting to Postgres",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:    "namespace",
				Aliases: []string{"n"},
			},
			&cli.StringFlag{
				Name:    "datebase",
				Aliases: []string{"d"},
			},
		},
		Subcommands: []*cli.Command{
			grantCommand(),
			passwordCommand(),
			prepareCommand(),
			proxyCommand(),
			psqlCommand(),
			revokeCommand(),
			usersCommand(),
		},
	}
}
