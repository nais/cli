package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func branchCommand(parent *flag.Postgres) *naistrix.Command {
	return &naistrix.Command{
		Name: "branch", Title: "Deprecated Postgres inspection alias.",
		Description: "Branching has been removed. Only the read-only list compatibility alias remains; use postgres get instead.",
		SubCommands: []*naistrix.Command{branchListCommand(parent)},
	}
}

func branchListCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchList{Postgres: parent}
	return &naistrix.Command{
		Name: "list", Title: "Show the deprecated main compatibility view.",
		Description: "Show the database as one main entry for script compatibility. There is no independently managed branch; prefer postgres get.",
		Args:        []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres database", ChoicesFunc: postgresNameChoices(parent)}}, Flags: f,
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			pg, err := postgres.GetPostgres(ctx, parent.Team, env, args.Get("postgres"))
			if err != nil {
				return err
			}
			main := postgres.Branch{Name: "main", State: pg.State}
			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render([]postgres.Branch{main})
			}
			printPostgresHeading(out, pg.Name, env)
			return out.Table().Render(branchListRows(postgres.BranchStatus{Active: "main", Branches: []postgres.Branch{main}}))
		},
	}
}

type branchListRow struct {
	Name   string
	Status postgres.State
	Active string
}

func branchListRows(status postgres.BranchStatus) []branchListRow {
	rows := make([]branchListRow, 0, len(status.Branches))
	for _, branch := range status.Branches {
		active := "No"
		if branch.Name == status.Active {
			active = "Yes"
		}
		rows = append(rows, branchListRow{Name: branch.Name, Status: postgres.State(branch.State), Active: active})
	}
	return rows
}

func branchEnvironment(ctx context.Context, f *flag.Postgres, name string) (string, error) {
	return resolvePostgresEnvironment(ctx, f.Team, name, string(f.Environment))
}
