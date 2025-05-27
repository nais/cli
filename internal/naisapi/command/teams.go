package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/output"
)

func teams(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Teams{Api: parentFlags}

	return cli.NewCommand("teams", "Get a list of your teams.",
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			teams, err := naisapi.GetUserTeams(ctx, flags)
			if err != nil {
				return err
			}

			if len(teams.Me.(*gql.UserTeamsMeUser).Teams.Nodes) == 0 {
				out.Println("No teams found.")
				return nil
			}

			for _, team := range teams.Me.(*gql.UserTeamsMeUser).Teams.Nodes {
				out.Println(team.Team.Slug, "-", team.Team.Purpose)
			}

			return nil
		}),
	)
}
