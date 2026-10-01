package command

import (
	"context"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func Postgres(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Postgres{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name: "postgres", Title: "Manage Nais Postgres instances (experimental).",
		Description: "Experimental commands for Nais Postgres branches and brokered personal access.", StickyFlags: flags,
		SubCommands: []*naistrix.Command{listCommand(flags), psqlCommand(flags), proxyCommand(flags)},
	}
}

func listCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.List{Postgres: parentFlags}
	return &naistrix.Command{
		Name: "list", Title: "List Nais Postgres branches for a team.",
		Description: "List Nais Postgres branches owned by a team.", Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			labelFilters, err := labels.ParseFilters(flags.Labels)
			if err != nil {
				return err
			}
			var environments []string
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
				out.Println("Team has no Nais Postgres branches.")
				return nil
			}
			return out.Table().Render(ret)
		},
	}
}
