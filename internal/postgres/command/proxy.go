package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/proxy"
)

func proxyCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Proxy{
		Postgres: parentFlags,
		Port:     5432,
		Host:     "localhost",
	}

	return cli.NewCommand("proxy", "Create a proxy to a SQL instance.",
		cli.WithLong("Allows your user to connect to databases and starts a proxy."),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return proxy.Run(ctx, args[0], flags)
		}),
		cli.WithFlag("port", "", "Port to use for the proxy.", &flags.Port),
		cli.WithFlag("host", "", "Host to use for the proxy.", &flags.Host),
	)
}
