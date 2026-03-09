package command

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

type SecretSummary struct {
	Name         string `heading:"Name"`
	Environment  string `heading:"Environment"`
	Keys         string `heading:"Keys"`
	Workloads    string `heading:"Workloads"`
	LastModified string `heading:"Last Modified"`
}

const maxListItems = 3

func list(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.List{Secret: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List secrets for a team.",
		Description: "This command lists all secrets for a given team.",
		Flags:       f,
		Examples: []naistrix.Example{
			{
				Description: "List all secrets for the team.",
			},
			{
				Description: "List secrets in a specific environment.",
				Command:     "--environment dev",
			},
			{
				Description: "List only secrets that are in use by workloads.",
				Command:     "--in-use",
			},
		},
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			secrets, err := secret.GetAll(ctx, f.Team)
			if err != nil {
				return fmt.Errorf("fetching secrets: %w", err)
			}

			if len(secrets) == 0 {
				out.Infoln("No secrets found")
				return nil
			}

			var summaries []SecretSummary
			for _, s := range secrets {
				envName := s.TeamEnvironment.Environment.Name

				if len(f.Environment) > 0 && !slices.Contains(f.Environment, envName) {
					continue
				}

				workloadCount := len(s.Workloads.Nodes)
				if f.InUse && workloadCount == 0 {
					continue
				}

				lastModified := ""
				if !s.LastModifiedAt.IsZero() {
					lastModified = s.LastModifiedAt.Format(time.DateTime)
				}

				var workloadNames []string
				for _, w := range s.Workloads.Nodes {
					workloadNames = append(workloadNames, w.GetName())
				}

				summaries = append(summaries, SecretSummary{
					Name:         s.Name,
					Environment:  envName,
					Keys:         summarizeList(s.Keys),
					Workloads:    summarizeList(workloadNames),
					LastModified: lastModified,
				})
			}

			if len(summaries) == 0 {
				out.Infoln("No secrets match the given filters")
				return nil
			}

			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(summaries)
			}

			return out.Table().Render(summaries)
		},
	}
}

// summarizeList joins items with ", " and truncates if there are more than maxListItems.
// e.g. ["a", "b", "c", "d", "e"] → "a, b, c, +2 more"
func summarizeList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	if len(items) <= maxListItems {
		return strings.Join(items, ", ")
	}

	return fmt.Sprintf("%s, +%d more", strings.Join(items[:maxListItems], ", "), len(items)-maxListItems)
}
