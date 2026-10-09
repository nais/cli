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

func revokeCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.Revoke{
		CloudSQL: parentFlags,
		Schema:   "public",
	}
	return &naistrix.Command{
		Name:  "revoke",
		Title: "Revoke database access granted by prepare.",
		Description: heredoc.Doc(`
			Revoke removes access for cloudsqliamuser by default, or for an existing Cloud SQL IAM group with --group.
			This uses the application credentials and affects tables, sequences and default privileges in the chosen schema.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name", Prompt: "Name of the application whose Cloud SQL database access should be revoked"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
				return err
			} else if !result {
				return fmt.Errorf("cancelled by user")
			}

			return cloudsql.RevokeAccess(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
		},
	}
}
