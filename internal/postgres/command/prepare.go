package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/prepare"
)

func prepareCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Prepare{
		Postgres: parentFlags,
		Schema:   "public",
	}

	return cli.NewCommand("prepare", "Prepare your SQL instance for use with personal accounts.",
		cli.WithLong(`Prepare will prepare the SQL instance by connecting using the
		 application credentials and modify the permissions on the public schema.
		 All IAM users in your GCP project will be able to connect to the instance.
		
		 This operation is only required to run once for each SQL instance.`),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return prepare.Run(ctx, args[0], flags)
		}),
		cli.WithFlag("all-privs", "", "Gives all privileges to users.", &flags.AllPrivileges),
		cli.WithFlag("schema", "", "Schema to grant access to.", &flags.Schema),
	)
}
