package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func list(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.ListApps{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "list",
		Title: "List applications in a team.",
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := app.GetTeamApplications(ctx, flags.Team, gql.ApplicationOrder{
				Field:     gql.ApplicationOrderFieldIssues,
				Direction: gql.OrderDirectionDesc,
			}, gql.TeamApplicationsFilter{Environments: flags.Environment})
			if err != nil {
				return err
			}

			return out.Table().Render(ret)
		},
	}
}
