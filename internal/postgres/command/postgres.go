package command

import (
	"context"

	"github.com/nais/cli/v2/internal/gcloud"
	"github.com/nais/cli/v2/internal/k8s"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func Postgres(parentFlags *root.Flags) *naistrix.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	flags := &flag.Postgres{
		Flags:     parentFlags,
		Namespace: flag.Namespace(defaultNamespace),
		Context:   flag.Context(defaultContext),
	}

	return &naistrix.Command{
		Name:        "postgres",
		Title:       "Manage SQL instances.",
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
		ValidateFunc: func(ctx context.Context, _ []string) error {
			_, err := gcloud.ValidateAndGetUserLogin(ctx, false)
			return err
		},
	}
}
