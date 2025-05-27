package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func passwordCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Password{Postgres: parentFlags}
	return cli.NewCommand("password", "Manage SQL instance passwords.",
		cli.WithSubCommands(
			cli.NewCommand("rotate", "Rotate the SQL instance password.",
				cli.WithLong("The rotation is done in GCP and in the Kubernetes secret."),
				cli.WithArgs("app_name"),
				cli.WithValidate(cli.ValidateExactArgs(1)),
				cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
					return postgres.RotatePassword(ctx, args[0], flags.Context, flags.Namespace, out)
				}),
			),
		),
	)
}
