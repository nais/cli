package command

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func Postgres(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Postgres{
		GlobalFlags: parentFlags,
	}

	return &naistrix.Command{
		Name:        "postgres",
		Title:       "Manage postgres instances.",
		Description: "Commands for managing Google Cloud SQL Postgres instances, including listing, migration, user management, password rotation, and direct database access.",
		Aliases:     []string{"pg"},
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listCommand(flags),
			migrateCommand(flags),
			passwordCommand(flags),
			usersCommand(flags),
			enableAuditCommand(flags),
			verifyAuditCommand(flags),
			grantCommand(flags),
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
