package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func passwordCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Password{Postgres: parentFlags}
	return &cli.Command{
		Name:        "password",
		Title:       "Manage SQL instance passwords.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
			{
				Name:        "rotate",
				Title:       "Rotate the SQL instance password.",
				Description: "The rotation is done in GCP and in the Kubernetes secret.",
				Args: []cli.Argument{
					{Name: "app_name", Required: true},
				},
				ValidateFunc: cli.ValidateExactArgs(1),
				RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
					return postgres.RotatePassword(ctx, args[0], flags.Context, flags.Namespace, out)
				},
			},
		},
	}
}
