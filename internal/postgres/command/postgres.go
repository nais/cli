package command

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func Postgres(parentFlags *flags.GlobalFlags) *naistrix.Command {
	defaultContext, _ := k8s.GetDefaultContextAndNamespace()
	flags := &flag.Postgres{
		GlobalFlags: parentFlags,
		Environment: flag.Environment(defaultContext),
	}

	return &naistrix.Command{
		Name:        "postgres",
		Title:       "Manage postgres instances.",
		Aliases:     []string{"pg"},
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
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
			if err := flags.UsesRemovedFlags(); err != nil {
				return err
			}
			_, err := gcloud.ValidateAndGetUserLogin(ctx, false)
			return err
		},
	}
}
