package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/root"
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
		Short:       "Manage SQL instances.",
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
			_, err := gcp.ValidateAndGetUserLogin(ctx, false)
			return err
		},
	}
}
