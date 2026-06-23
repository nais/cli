package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/status"
	"github.com/nais/cli/internal/status/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

type workload struct {
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Environment string   `json:"environment"`
	Messages    []string `json:"messages"`
}

type workloadsWithIssues []workload

func (f workloadsWithIssues) String() string {
	if len(f) == 0 {
		return "No issues detected"
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "%v workloads with issues\n\n", len(f))
	for _, w := range f {
		_, _ = fmt.Fprintf(&b, "%s (%s): %s\n", w.Kind, w.Environment, w.Name)
		_, _ = fmt.Fprintf(&b, "%s\n\n", strings.Join(w.Messages, "\n"))
	}

	return strings.TrimRight(b.String(), "\n")
}

type statusEntry struct {
	Team      output.Link         `json:"team"`
	Workloads int                 `json:"workloads"`
	NotNais   int                 `heading:"Not Nais" json:"notNais"`
	Issues    workloadsWithIssues `heading:"Critical Issues" json:"failing"`
}

func Status(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Status{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "status",
		Title:       "Get a quick overview of the status of your teams.",
		Description: "Show the status of your teams, including workload counts and critical issues such as missing instances, vulnerabilities, and failed job runs.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			ret, err := status.GetStatus(ctx, flags)
			if err != nil {
				return err
			}

			critical := gql.SeverityCritical
			var entries []statusEntry
			for _, t := range ret {
				teamIssues, err := issues.GetAll(ctx, t.Team.Slug, gql.IssueFilter{Severity: critical})
				if err != nil {
					return err
				}

				// Group issues by resource (name + environment).
				type resourceKey struct{ name, env string }
				resourceMap := make(map[resourceKey]*workload)
				var resourceOrder []resourceKey
				for _, issue := range teamIssues {
					key := resourceKey{issue.ResourceName, issue.Environment}
					if _, ok := resourceMap[key]; !ok {
						resourceMap[key] = &workload{
							Kind:        issue.ResourceType,
							Name:        issue.ResourceName,
							Environment: issue.Environment,
						}
						resourceOrder = append(resourceOrder, key)
					}
					resourceMap[key].Messages = append(resourceMap[key].Messages, issue.Message)
				}

				n := statusEntry{
					Team: output.Link{
						Name: t.Team.Slug,
						URL:  fmt.Sprintf("https://%s/team/%s", user.ConsoleHost(), t.Team.Slug),
					},
					Workloads: t.Team.Workloads.PageInfo.TotalCount,
					NotNais:   len(resourceMap),
					Issues:    make(workloadsWithIssues, 0, len(resourceOrder)),
				}
				for _, key := range resourceOrder {
					n.Issues = append(n.Issues, *resourceMap[key])
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
