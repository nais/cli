package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func listCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.List{Postgres: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List postgres instances for a team.",
		Description: "List all Google Cloud SQL Postgres instances owned by a team, showing instance details.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := postgres.GetTeamPostgresInstances(ctx, flags.Team, nil)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("Team has no postgres instances.")
				return nil
			}

			return out.Table().Render(ret)
		},
	}
}
