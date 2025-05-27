package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/revoke"
)

func revokeCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Revoke{
		Postgres: parentFlags,
		Schema:   "public",
	}

	return cli.NewCommand("revoke", "Revoke access to your SQL instance for the role 'cloudsqliamuser'.",
		cli.WithLong(`Revoke will revoke the role 'cloudsqliamuser' access to the tables in the SQL instance.

This is done by connecting using the application credentials and modify the permissions on the public schema.

 This operation is only required to run once for each SQL instance.`),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return revoke.Run(ctx, args[0], flags)
		}),
		cli.WithFlag("schema", "", "Schema to revoke access from.", &flags.Schema),
	)
}
