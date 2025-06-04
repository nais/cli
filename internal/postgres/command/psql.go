package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func psqlCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Psql{Postgres: parentFlags}
	return &cli.Command{
		Name:  "psql",
		Short: "Connect to the database using psql.",
		Long:  "Create a shell to the SQL instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		RunFunc: func(ctx context.Context, out output.Output, args []string) error {
			return postgres.RunPSQL(ctx, args[0], flags.Context, flags.Namespace, flags.IsVerbose(), out)
		},
	}
}
