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
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	flags := &flag.Postgres{
		GlobalFlags: parentFlags,
		Namespace:   flag.Namespace(defaultNamespace),
		Context:     flag.Context(defaultContext),
	}

	return &naistrix.Command{
		Name:        "postgres",
		Title:       "Manage postgres instances.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			migrateCommand(flags),
			passwordCommand(flags),
			usersCommand(flags),
			enableAuditCommand(flags),
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
