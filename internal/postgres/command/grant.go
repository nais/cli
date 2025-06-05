package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func grantCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Grant{Postgres: parentFlags}
	return &cli.Command{
		Name:  "grant",
		Short: "Grant yourself access to a SQL instance database.",
		Long:  "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		Flags:        flags,
		RunFunc: func(ctx context.Context, out output.Output, args []string) error {
			return postgres.GrantAndCreateSQLUser(ctx, args[0], flags.Context, flags.Namespace, out)
		},
	}
}
