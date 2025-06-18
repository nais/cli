package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
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
			{Name: "app_name", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		Flags:        flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return postgres.RunProxy(ctx, args[0], flags.Context, flags.Namespace, flags.Host, flags.Port, flags.IsVerbose(), out)
		},
	}
}
