package command

import (
	"context"
	"fmt"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func revokeCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Revoke{
		Postgres: parentFlags,
		Schema:   "public",
	}

	return &naistrix.Command{
		Name:  "revoke",
		Title: `Revoke access to your SQL instance for the role "cloudsqliamuser".`,
		Description: heredoc.Doc(`
			Revoke will revoke the role "cloudsqliamuser" access to the tables in the SQL instance.

			This is done by connecting using the application credentials and modify the permissions on the public schema.

			This operation is only required to run once for each SQL instance.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			return postgres.RevokeAccess(ctx, args.Get("app_name"), flags, out)
		},
	}
}
