package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/gcloud"
	"github.com/nais/cli/v2/internal/k8s"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
	"github.com/nais/cli/v2/internal/root"
)

func Postgres(parentFlags *root.Flags) *cli.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	flags := &flag.Postgres{
		Flags:     parentFlags,
		Namespace: flag.Namespace(defaultNamespace),
		Context:   flag.Context(defaultContext),
	}

	return &cli.Command{
		Name:        "postgres",
		Title:       "Manage SQL instances.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
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
