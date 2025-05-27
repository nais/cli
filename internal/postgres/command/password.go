package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/password/rotate"
)

func passwordCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Password{Postgres: parentFlags}
	return cli.NewCommand("password", "Manage SQL instance passwords.",
		cli.WithSubCommands(
			cli.NewCommand("rotate", "Rotate the SQL instance password.",
				cli.WithLong("The rotation is done in GCP and in the Kubernetes secret."),
				cli.WithArgs("app_name"),
				cli.WithValidate(cli.ValidateExactArgs(1)),
				cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
					return rotate.Run(ctx, args[0], &flag.PasswordRotate{Password: flags})
				}),
			),
		),
	)
}
