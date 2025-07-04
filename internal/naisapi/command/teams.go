package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/writer"
	"github.com/savioxavier/termlink"
)

func teams(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Teams{
		Api:    parentFlags,
		Output: "table",
	}

	return &naistrix.Command{
		Name:  "teams",
		Title: "Get a list of your teams.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
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

			if len(teams) == 0 {
				out.Println("No teams found.")
				return nil
			}

			var w writer.Writer
			if flags.Output == "json" {
				w = writer.NewJSON(out, true)
			} else {
				tbl := writer.NewTable(out, writer.WithColumns("Slug", "Description"), writer.WithFormatter(func(row, column int, value any) string {
					if column != 0 {
						return fmt.Sprint(value)
					}

					slug := fmt.Sprint(value)
					return termlink.ColorLink(slug, "https://console.nav.cloud.nais.io/team/"+slug, "underline")
				}))
				w = tbl
			}

			return w.Write(teams)
		},
	}
}
