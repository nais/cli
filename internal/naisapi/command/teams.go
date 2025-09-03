package command

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func teamsCommand(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Teams{
		Api:    parentFlags,
		Output: "table",
	}

	return &naistrix.Command{
		Name:  "teams",
		Title: "Get a list of your teams.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// TODO: Once https://github.com/pterm/pterm/issues/697 is resolved, we can use a link to Console instead of just the slug.
			type team struct {
				Slug        string `json:"slug"`
				Description string `json:"description"`
			}

			var teams []team

			if flags.All {
				ret, err := naisapi.GetAllTeams(ctx)
				if err != nil {
					return err
				}

				for _, t := range ret.Teams.Nodes {
					teams = append(teams, team{
						Slug:        t.Slug,
						Description: t.Purpose,
					})
				}
			} else {
				userTeams, err := naisapi.GetUserTeams(ctx)
				if err != nil {
					return err
				}

				for _, t := range userTeams {
					teams = append(teams, team{
						Slug:        t.Team.Slug,
						Description: t.Team.Purpose,
					})
				}
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(teams)
			}

			if len(teams) == 0 {
				out.Println("No teams found.")
				return nil
			}

			return out.Table().Render(teams)
		},
	}
}
