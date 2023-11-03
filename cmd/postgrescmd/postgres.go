package postgrescmd

import (
	"github.com/nais/cli/pkg/gcp"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "postgres",
		Usage: "Command used for connecting to Postgres",
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
