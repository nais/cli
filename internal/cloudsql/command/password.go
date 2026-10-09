package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func passwordCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.Password{CloudSQL: parentFlags}
	return &naistrix.Command{
		Name:        "password",
		Title:       "Manage SQL instance passwords.",
		Description: "Commands for managing Cloud SQL instance passwords, including password rotation.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			{
				Name:        "rotate",
				Title:       "Rotate the SQL instance password.",
				Description: "The rotation is done in GCP and in the Kubernetes secret.",
				Args: []naistrix.Argument{
					{Name: "app_name", Prompt: "Name of the application whose Cloud SQL password should be rotated"},
				},
				ValidateFunc: validation.RequireTeamAndEnvironment(flags),
				RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
					return cloudsql.RotatePassword(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
				},
			},
		},
	}
}
