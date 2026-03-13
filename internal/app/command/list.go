package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/savioxavier/termlink"
)

type appName struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (a appName) String() string {
	return termlink.Link(a.Name, a.URL)
}

func list(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.List{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "list",
		Title: "List applications in a team.",
		Flags: flags,
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if flags.Team == "" {
					return nil, "Please provide team. 'nais config set team <team>', or '--team <team>' flag."
				}
			}
			return nil, ""
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := app.GetTeamApplications(ctx, flags.Team, gql.ApplicationOrder{
				Field:     gql.ApplicationOrderFieldIssues,
				Direction: gql.OrderDirectionDesc,
			}, gql.TeamApplicationsFilter{Environments: flags.Environment})
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			type entry struct {
				State         app.State          `json:"state"`
				Name          appName            `json:"name"`
				Environment   string             `json:"environment"`
				InstancesInfo *app.InstancesInfo `heading:"Running" json:"running"`
				IssueInfo     *app.IssueInfo     `heading:"Issues" json:"issue_info"`
				LastUpdated   app.LastUpdated    `heading:"Last Updated" json:"last_updated"`
			}

			entries := make([]entry, 0, len(ret))
			for _, a := range ret {
				entries = append(entries, entry{
					State: a.State,
					Name: appName{
						Name: a.Name,
						URL: fmt.Sprintf(
							"https://%s/team/%s/%s/app/%s",
							user.ConsoleHost(),
							flags.Team,
							a.Environment,
							a.Name,
						),
					},
					Environment:   a.Environment,
					InstancesInfo: a.InstancesInfo,
					IssueInfo:     a.IssueInfo,
					LastUpdated:   a.LastUpdated,
				})
			}

			return out.Table().Render(entries)
		},
	}
}
