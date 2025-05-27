package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/psql"
)

func psqlCommand(parentFlags *flag.Postgres) *cli.Command {
	return cli.NewCommand("psql", "Connect to the database using psql.",
		cli.WithLong("Create a shell to the SQL instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell."),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return psql.Run(ctx, args[0], &flag.Psql{Postgres: parentFlags})
		}),
	)
}
