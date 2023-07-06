package postgrescmd

import (
	"github.com/nais/cli/pkg/gcp"
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
				Name:    "database",
				Aliases: []string{"d"},
			},
		},
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context, false)
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
