package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/postgres"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
)

func proxyCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Proxy{
		Postgres: parentFlags,
		Port:     5432,
		Host:     "localhost",
	}

	return &cli.Command{
		Name:        "proxy",
		Title:       "Create a proxy to a SQL instance.",
		Description: "Allows your user to connect to databases and starts a proxy.",
		Args: []cli.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return postgres.RunProxy(ctx, args[0], flags.Context, flags.Namespace, flags.Host, flags.Port, flags.IsVerbose(), out)
		},
	}
}
