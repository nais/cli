package command

import (
	"context"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/postgres/command/migrate"
	"github.com/urfave/cli/v3"
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
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			_, err := gcp.ValidateAndGetUserLogin(ctx, false)
			return ctx, err
		},
		Commands: commands,
	}
}
