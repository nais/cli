package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/savioxavier/termlink"
	"k8s.io/utils/strings/slices"
)

type teamWorkload struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (tw teamWorkload) String() string {
	return termlink.Link(tw.Name, tw.Url)
}

func teamCommand(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Team{Api: parentFlags}
	return &naistrix.Command{
		Name:  "team",
		Title: "Operations on a team.",
		SubCommands: []*naistrix.Command{
			listWorkloads(flags),
		},
	}
}

func listWorkloads(parentFlags *flag.Team) *naistrix.Command {
	flags := &flag.ListWorkloads{
		Team:   parentFlags,
		Output: "table",
	}

	return &naistrix.Command{
		Name:  "list-workloads",
		Title: "List workloads of a team.",
		Args: []naistrix.Argument{
			{Name: "team"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			type entry struct {
				Workload        teamWorkload `json:"workload"`
				Environment     string       `json:"environment"`
				Type            string       `json:"type"`
				State           string       `json:"state"`
				Vulnerabilities int          `json:"vulnerabilities"`
				Issues          int          `heading:"Critical Issues" json:"issues"`
			}

			teamSlug := args.Get("team")
			ret, err := naisapi.GetTeamWorkloads(ctx, teamSlug)
			if err != nil {
				return err
			}

			entries := make([]entry, len(ret))
			for i, w := range ret {
				state := "(unknown)"
				switch actual := w.(type) {
				case *gql.GetTeamWorkloadsTeamWorkloadsWorkloadConnectionNodesApplication:
					state = string(actual.GetApplicationState())
				case *gql.GetTeamWorkloadsTeamWorkloadsWorkloadConnectionNodesJob:
					state = string(actual.GetJobState())
				}

				workloadType := "app"
				if w.GetTypename() == "Job" {
					workloadType = "job"
				}

				entries[i] = entry{
					Workload: teamWorkload{
						Name: w.GetName(),
						Url: fmt.Sprintf(
							"https://%s/team/%s/%s/%s/%s",
							user.ConsoleHost(),
							teamSlug,
							w.GetTeamEnvironment().Environment.Name,
							workloadType,
							w.GetName(),
						),
					},
					Environment:     w.GetTeamEnvironment().Environment.Name,
					Type:            w.GetTypename(),
					State:           state,
					Vulnerabilities: w.GetImage().VulnerabilitySummary.Total,
					Issues:          w.GetTotalIssues().PageInfo.TotalCount,
				}
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(entries)
			}

			if len(ret) == 0 {
				out.Println("Team has no workloads.")
				return nil
			}

			return out.Table().Render(entries)
		},
		AutoCompleteFunc: func(ctx context.Context, _ *naistrix.Arguments, toComplete string) ([]string, string) {
			if len(toComplete) < 2 {
				return nil, "Provide at least 2 characters to auto-complete team slugs."
			}

			slugs, err := naisapi.GetAllTeamSlugs(ctx)
			if err != nil {
				return nil, "Unable to fetch team slugs."
			}

			return slices.Filter([]string{}, slugs, func(slug string) bool {
				return strings.HasPrefix(slug, toComplete)
			}), "Choose a team to list the workloads of."
		},
	}
}
