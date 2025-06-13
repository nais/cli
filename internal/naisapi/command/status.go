package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/cli/writer"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/savioxavier/termlink"
)

func status(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Status{Api: parentFlags}
	return &cli.Command{
		Name:  "status",
		Short: "Get a quick overview of the status of your teams.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			type failing struct {
				Kind        string   `json:"kind"`
				Name        string   `json:"name"`
				Environment string   `json:"environment"`
				ErrorTypes  []string `json:"errorType"`
			}

			type team struct {
				Slug    string    `json:"slug"`
				Total   int       `json:"total"`
				NotNais int       `json:"notNais"`
				Failing []failing `json:"failing"`
			}

			var teams []team

			ret, err := naisapi.GetStatus(ctx, flags)
			if err != nil {
				return err
			}

			for _, t := range ret.Me.(*gql.TeamStatusMeUser).Teams.Nodes {
				n := team{
					Slug:    t.Team.Slug,
					Total:   t.Team.Total.PageInfo.TotalCount,
					NotNais: t.Team.NotNice.PageInfo.TotalCount,
					Failing: []failing{},
				}
				for _, f := range t.Team.Failing.Nodes {
					a := failing{
						Kind:        f.GetTypename(),
						Name:        f.GetName(),
						Environment: f.GetTeamEnvironment().Environment.Name,
					}
					for _, et := range f.GetStatus().Errors {
						a.ErrorTypes = append(a.ErrorTypes, et.GetTypename())
					}
					n.Failing = append(n.Failing, a)
				}
				teams = append(teams, n)
			}

			if len(teams) == 0 {
				out.Println("No teams found.")
				return nil
			}

			var w writer.Writer
			if flags.Output == "json" {
				w = writer.NewJSON(out, true)
			} else {
				tbl := writer.NewTable(out, writer.WithColumns("Slug", "Total", "Not Nais", "Failing"), writer.WithFormatter(func(row, column int, value any) string {
					switch column {
					case 0:
						slug := fmt.Sprint(value)
						return termlink.ColorLink(slug, "https://console.nav.cloud.nais.io/team/"+slug, "underline")
					case 3:
						failing := value.([]failing)
						if len(failing) == 0 {
							return "No failing workloads"
						}
						var failingStr string
						for _, f := range failing {
							failingStr += fmt.Sprintf("%s (%s): %s\n", f.Kind, f.Environment, f.Name)
							if len(f.ErrorTypes) > 0 {
								failingStr += fmt.Sprint(f.ErrorTypes) + "\n"
							}
						}
					}
					return fmt.Sprint(value)
				}))
				w = tbl
			}

			return w.Write(teams)
		},
	}
}
