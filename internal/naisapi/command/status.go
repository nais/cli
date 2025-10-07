package command

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/savioxavier/termlink"
)

type team struct {
	Slug string `json:"slug"`
	Url  string `json:"url"`
}

func (t team) String() string {
	return termlink.Link(t.Slug, t.Url)
}

type workload struct {
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Environment string   `json:"environment"`
	ErrorTypes  []string `json:"errorType"`
}

type workloadsWithIssues []workload

func (f workloadsWithIssues) String() string {
	if len(f) == 0 {
		return "No issues detected"
	}

	issues := fmt.Sprintf("%v workloads with issues\n\n", len(f))
	for _, w := range f {
		issues += fmt.Sprintf("%s (%s): %s\n", w.Kind, w.Environment, w.Name)
		issues += formatErrorTypes(w.ErrorTypes) + "\n\n"
	}

	return strings.TrimRight(issues, "\n")
}

type statusEntry struct {
	Team      team                `json:"team"`
	Workloads int                 `json:"workloads"`
	NotNais   int                 `heading:"Not Nais" json:"notNais"`
	Issues    workloadsWithIssues `heading:"Critical Issues" json:"failing"`
}

func statusCommand(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Status{Api: parentFlags}
	return &naistrix.Command{
		Name:  "status",
		Title: "Get a quick overview of the status of your teams.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			ret, err := naisapi.GetStatus(ctx, flags)
			if err != nil {
				return err
			}

			var entries []statusEntry
			for _, t := range ret {
				workloadsWithCriticalIssues := make([]gql.TeamStatusMeUserTeamsTeamMemberConnectionNodesTeamMemberTeamWorkloadsWorkloadConnectionNodesWorkload, 0)
				for _, w := range t.Team.Workloads.Nodes {
					if w.GetIssues().PageInfo.TotalCount > 0 {
						workloadsWithCriticalIssues = append(workloadsWithCriticalIssues, w)
					}
				}

				n := statusEntry{
					Team: team{
						Slug: t.Team.Slug,
						Url:  fmt.Sprintf("https://%s/team/%s", user.ConsoleHost(), t.Team.Slug),
					},
					Workloads: t.Team.Workloads.PageInfo.TotalCount,
					NotNais:   len(workloadsWithCriticalIssues),
					Issues:    make(workloadsWithIssues, 0),
				}
				for _, f := range workloadsWithCriticalIssues {
					a := workload{
						Kind:        f.GetTypename(),
						Name:        f.GetName(),
						Environment: f.GetTeamEnvironment().Environment.Name,
					}
					for _, et := range f.GetIssues().Nodes {
						a.ErrorTypes = append(a.ErrorTypes, et.GetTypename())
					}
					n.Issues = append(n.Issues, a)
				}
				entries = append(entries, n)
			}

			if len(entries) == 0 {
				out.Println("No teams found.")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(entries)
			}

			return out.Table().Render(entries)
		},
	}
}

func formatErrorTypes(errorTypes []string) string {
	if len(errorTypes) == 0 {
		return "Unknown failure"
	}

	texts := map[string]string{}
	for _, et := range errorTypes {
		switch et {
		case "WorkloadStatusNoRunningInstances":
			texts[et] = "No running instances"
		case "WorkloadStatusVulnerable":
			texts[et] = "Vulnerabilities detected"
		case "WorkloadStatusFailedRun":
			texts[et] = "Failed job run"
		default:
			texts[et] = et
		}
	}

	vals := maps.Values(texts)
	ret := slices.Collect(vals)
	slices.Sort(ret)

	return strings.Join(ret, "\n")
}
