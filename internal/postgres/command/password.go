package command

import (
	"context"

	"github.com/nais/cli/v2/internal/postgres"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func passwordCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Password{Postgres: parentFlags}
	return &naistrix.Command{
		Name:        "password",
		Title:       "Manage SQL instance passwords.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			{
				Name:        "rotate",
				Title:       "Rotate the SQL instance password.",
				Description: "The rotation is done in GCP and in the Kubernetes secret.",
				Args: []naistrix.Argument{
					{Name: "app_name"},
				},
				RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
					return postgres.RotatePassword(ctx, args[0], flags.Context, flags.Namespace, out)
				},
			},
		},
	}
}
