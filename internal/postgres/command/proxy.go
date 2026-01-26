package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
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
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
				return postgres.RunProxy(ctx, args.Get("app_name"), flags.Context, flags.Namespace, flags.Host, flags.Port, flags.IsVerbose(), flags.Reason, out)
			},
	}
}
