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
		SubCommands: []*naistrix.Command{listCommand(flags), instanceGetCommand(flags), instanceCreateCommand(flags), instanceUpdateCommand(flags), instanceDeleteCommand(flags), branchCommand(flags), psqlCommand(flags), proxyCommand(flags)},
	}
}

func listCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.List{Postgres: parentFlags}
	return &naistrix.Command{
		Name: "list", Title: "List Nais Postgres databases for a team.",
		Description: "List Nais Postgres databases owned by a team. Use 'postgres branch list <postgres>' to list branches.", Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			labelFilters, err := labels.ParseFilters(flags.Labels)
			if err != nil {
				return err
			}
			var environments []string
			if flags.Environment != "" {
				environments = []string{string(flags.Environment)}
			}
			ret, err := postgres.GetTeamPostgreses(ctx, flags.Team, environments, labelFilters)
			if err != nil {
				return err
			}
			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}
			if len(ret) == 0 {
				if len(environments) > 0 || len(labelFilters) > 0 {
					out.Println("No Postgres databases match these filters.")
				} else {
					out.Println("No Postgres databases found.")
				}
				return nil
			}
			return out.Table().Render(ret)
		},
	}
}
