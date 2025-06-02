package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/cli/writer"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/output"
)

func teams(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Teams{
		Api:    parentFlags,
		Output: "table",
	}

	return cli.NewCommand("teams", "Get a list of your teams.",
		cli.WithFlag("all", "a", "List all teams, not just the ones you are a member of", &flags.All),
		cli.WithFlag("output", "o", "Format output (table|json)", &flags.Output),
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
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
				ret, err := naisapi.GetUserTeams(ctx, flags)
				if err != nil {
					return err
				}

				for _, t := range ret.Me.(*gql.UserTeamsMeUser).Teams.Nodes {
					teams = append(teams, team{
						Slug:        t.Team.Slug,
						Description: t.Team.Purpose,
					})
				}
			}

			if len(teams) == 0 {
				out.Println("No teams found.")
				return nil
			}

			var w writer.Writer
			if flags.Output == "json" {
				w = writer.NewJSON(out, true)
			} else {
				tbl := writer.NewTable(out)
				tbl.AddColumn("Slug", "Slug")
				tbl.AddColumn("Description", "Description")
				w = tbl
			}

			return w.Write(teams)
		}),
	)
}
