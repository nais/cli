package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/naistrix"
)

func CloudSQL(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.CloudSQL{
		GlobalFlags: parentFlags,
	}

	return &naistrix.Command{
		Name: "cloudsql",
		// TODO: Remove the aliases once users have moved from the old `nais postgres` (Cloud SQL) commands.
		Aliases:     []string{"postgres", "pg"},
		Title:       "Manage Google Cloud SQL instances.",
		Description: "Manage Google Cloud SQL instances, including listing, migration, user management, password rotation, and direct database access.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listCommand(flags),
			migrateCommand(flags),
			passwordCommand(flags),
			usersCommand(flags),
			enableAuditCommand(flags),
			verifyAuditCommand(flags),
			prepareCommand(flags),
			proxyCommand(flags),
			psqlCommand(flags),
			revokeCommand(flags),
		},
		ValidateFunc: func(ctx context.Context, _ *naistrix.Arguments) error {
			_, err := gcloud.ValidateAndGetUserLogin(ctx, false)
			return err
		},
	}
}
