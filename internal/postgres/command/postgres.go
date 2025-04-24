package command

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/postgres/command/migrate"
	"github.com/urfave/cli/v2"
)

func Postgres() *cli.Command {
	commands := []*cli.Command{
		audit(),
		grant(),
		migrate.Migrate(),
		password(),
		prepare(),
		proxy(),
		psql(),
		revoke(),
		users(),
	}

	return &cli.Command{
		Name:  "postgres",
		Usage: "Command used for connecting to Postgres",
		Before: func(context *cli.Context) error {
			_, err := gcp.ValidateAndGetUserLogin(context.Context, false)
			return err
		},
		Subcommands: commands,
	}
}
