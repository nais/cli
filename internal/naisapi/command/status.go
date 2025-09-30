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
)

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

type team struct {
	// TODO: Once https://github.com/pterm/pterm/issues/697 is resolved, we can use a link to Console instead of just the slug.
	Slug      string              `json:"slug"`
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
			var teams []team

			ret, err := naisapi.GetStatus(ctx, flags)
			if err != nil {
				return err
			}

			for _, t := range ret {
				workloadsWithCriticalIssues := make([]gql.TeamStatusMeUserTeamsTeamMemberConnectionNodesTeamMemberTeamWorkloadsWorkloadConnectionNodesWorkload, 0)
				for _, w := range t.Team.Workloads.Nodes {
					if w.GetIssues().PageInfo.TotalCount > 0 {
						workloadsWithCriticalIssues = append(workloadsWithCriticalIssues, w)
					}
				}

				n := team{
					Slug:      t.Team.Slug,
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
				teams = append(teams, n)
			}

			if len(teams) == 0 {
				out.Println("No teams found.")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(teams)
			}

			return out.Table().Render(teams)
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
