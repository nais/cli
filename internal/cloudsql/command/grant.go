package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func grantCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.Grant{CloudSQL: parentFlags}
	return &naistrix.Command{
		Name:  "grant",
		Title: "Grant yourself access to a SQL instance database.",
		Description: heredoc.Doc(`
			This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.

			Not needed if you are a member of a Cloud SQL IAM group on the instance; use prepare --group to grant access to the group instead.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return cloudsql.GrantAndCreateSQLUser(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), out)
		},
	}
}
