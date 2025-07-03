package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func grantCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Grant{Postgres: parentFlags}
	return &naistrix.Command{
		Name:        "grant",
		Title:       "Grant yourself access to a SQL instance database.",
		Description: "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return postgres.GrantAndCreateSQLUser(ctx, args[0], flags.Context, flags.Namespace, out)
		},
	}
}
