package postgrescmd

import (
	"github.com/nais/cli/cmd/postgrescmd/migratecmd"
	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/metrics"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	commands := []*cli.Command{
		grantCommand(),
		migratecmd.Command(),
		passwordCommand(),
		prepareCommand(),
		proxyCommand(),
		psqlCommand(),
		revokeCommand(),
		usersCommand(),
	}

	return &cli.Command{
		Name:  "postgres",
		Usage: "Command used for connecting to Postgres",
		Before: func(context *cli.Context) error {
			metrics.AddOne("postgres_connect_total")
			return gcp.ValidateUserLogin(context.Context, false)
		},
		Subcommands: commands,
	}
}
