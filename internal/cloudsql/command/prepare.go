package command

import (
	"context"
	"fmt"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func prepareCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.Prepare{
		CloudSQL: parentFlags,
		Schema:   "public",
	}

	return &naistrix.Command{
		Name:  "prepare",
		Title: "Prepare your SQL instance for use with personal accounts.",
		Description: heredoc.Doc(`
			Prepare grants access to tables and sequences in the chosen schema using the application credentials.
			By default, access is granted to cloudsqliamuser (individual IAM database users).
			Use --group to grant access to an existing Cloud SQL IAM group instead.
			Group membership determines who inherits the group's database privileges.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name", Prompt: "Name of the application whose Cloud SQL database should have access prepared"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
				return err
			} else if !result {
				return fmt.Errorf("cancelled by user")
			}

			return cloudsql.PrepareAccess(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
		},
	}
}
