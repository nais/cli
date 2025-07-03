package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func psqlCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Psql{Postgres: parentFlags}
	return &naistrix.Command{
		Name:        "psql",
		Title:       "Connect to the database using psql.",
		Description: "Create a shell to the SQL instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return postgres.RunPSQL(ctx, args[0], flags.Context, flags.Namespace, flags.IsVerbose(), out)
		},
	}
}
