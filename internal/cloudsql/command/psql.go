package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func psqlCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.Psql{CloudSQL: parentFlags}
	return &naistrix.Command{
		Name:        "psql",
		Title:       "Connect to the database using psql.",
		Description: "Create a shell to the SQL instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return cloudsql.RunPSQL(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
		},
	}
}
