package command

import (
	"context"
	"fmt"

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
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			type entry struct {
				Team        team   `json:"team"`
				Description string `json:"description"`
			}

			var entries []entry

			if flags.All {
				ret, err := naisapi.GetAllTeams(ctx)
				if err != nil {
					return err
				}

				for _, t := range ret.Teams.Nodes {
					entries = append(entries, entry{
						Team: team{
							Slug: t.Slug,
							Url:  fmt.Sprintf("https://%s/team/%s", user.ConsoleHost(), t.Slug),
						},
						Description: t.Purpose,
					})
				}
			} else {
				userTeams, err := naisapi.GetUserTeams(ctx)
				if err != nil {
					return err
				}

				for _, t := range userTeams {
					entries = append(entries, entry{
						Team: team{
							Slug: t.Team.Slug,
							Url:  fmt.Sprintf("https://%s/team/%s", user.ConsoleHost(), t.Team.Slug),
						},
						Description: t.Team.Purpose,
					})
				}
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(entries)
			}

			if len(entries) == 0 {
				out.Println("No teams found.")
				return nil
			}

			return out.Table().Render(entries)
		},
	}
}
