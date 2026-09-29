package command

import (
	"context"

	"github.com/nais/cli/internal/labels"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func listCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.List{Postgres: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List Postgres branches and Cloud SQL instances for a team.",
		Description: "List NAIS Postgres branches and Google Cloud SQL Postgres instances owned by a team.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			labelFilters, err := labels.ParseFilters(flags.Labels)
			if err != nil {
				return err
			}

			environments := []string(nil)
			if flags.Environment != "" {
				environments = []string{string(flags.Environment)}
			}

			ret, err := postgres.GetTeamPostgresBranches(ctx, flags.Team, environments, labelFilters)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("Team has no Postgres branches or Cloud SQL instances.")
				return nil
			}

			return out.Table().Render(ret)
		},
	}
}
