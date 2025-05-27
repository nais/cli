package command

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/root"
)

func Postgres(parentFlags *root.Flags) *cli.Command {
	/*
		TODO: Enable support for this

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			_, err := gcp.ValidateAndGetUserLogin(cmd.Context(), false)
			return err
		},
	*/
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()

	flags := &flag.Postgres{
		Flags:     parentFlags,
		Namespace: defaultNamespace,
		Context:   defaultContext,
	}

	return cli.NewCommand("postgres", "Manage SQL instances.",
		cli.WithStickyFlag("namespace", "n", "The kubernetes `NAMESPACE` to use.", &flags.Namespace),
		cli.WithStickyFlag("context", "c", "The kubeconfig `CONTEXT` to use.", &flags.Context),
		cli.WithSubCommands(
			migrateCommand(flags),
			passwordCommand(flags),
			usersCommand(flags),
			enableAuditCommand(flags),
			grantCommand(flags),
			prepareCommand(flags),
			proxyCommand(flags),
			psqlCommand(flags),
			revokeCommand(flags),
		),
	)
}
