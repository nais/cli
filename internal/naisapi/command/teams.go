package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/cli/writer"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

func teams(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Teams{
		Api:    parentFlags,
		Output: "table",
	}

	return &cli.Command{
		Name:  "teams",
		Short: "Get a list of your teams.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
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
				tbl := writer.NewTable(out, writer.WithColumns("Slug", "Description"))
				w = tbl
			}

			return w.Write(teams)
		},
	}
}
