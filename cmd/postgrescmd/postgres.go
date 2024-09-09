package postgrescmd

import (
	"github.com/nais/cli/cmd/postgrescmd/migratecmd"
	"github.com/nais/cli/pkg/gcp"
	"github.com/urfave/cli/v2"
	"os"
)

func Command() *cli.Command {
	commands := []*cli.Command{
		grantCommand(),
		passwordCommand(),
		prepareCommand(),
		proxyCommand(),
		psqlCommand(),
		revokeCommand(),
		usersCommand(),
	}

	// Poor mans feature flag for enabling migrate command
	if os.Getenv("NAIS_ENABLE_MIGRATE") == "true" {
		commands = append(commands, migratecmd.Command())
	}

	return &cli.Command{
		Name:  "postgres",
		Usage: "Command used for connecting to Postgres",
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context, false)
		},
		Subcommands: commands,
	}
}
