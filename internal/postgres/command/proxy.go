package command

import (
	"context"

	"github.com/nais/cli/v2/internal/postgres"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func proxyCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Proxy{
		Postgres: parentFlags,
		Port:     5432,
		Host:     "localhost",
	}

	return &naistrix.Command{
		Name:        "proxy",
		Title:       "Create a proxy to a SQL instance.",
		Description: "Allows your user to connect to databases and starts a proxy.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return postgres.RunProxy(ctx, args[0], flags.Context, flags.Namespace, flags.Host, flags.Port, flags.IsVerbose(), out)
		},
	}
}
