package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/grant"
)

func grantCommand(parentFlags *flag.Postgres) *cli.Command {
	return cli.NewCommand("grant", "Grant yourself access to a SQL instance database.",
		cli.WithLong("This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email."),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return grant.Run(ctx, args[0], &flag.Grant{Postgres: parentFlags})
		}),
	)
}
