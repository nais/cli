package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

type ConfigSummary struct {
	Name         string              `heading:"Name"`
	Environment  string              `heading:"Environment"`
	Keys         string              `heading:"Keys"`
	Workloads    string              `heading:"Workloads"`
	LastModified config.LastModified `heading:"Last Modified"`
}

const maxListItems = 3

func list(parentFlags *flag.Config) *naistrix.Command {
	f := &flag.List{Config: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List config for a team.",
		Description: "This command lists all config for a given team.",
		Flags:       f,
		Examples: []naistrix.Example{
			{
				Description: "List all config for the team.",
			},
			{
				Description: "List config in a specific environment.",
				Command:     "--environment dev",
			},
		},
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			configs, err := config.GetAll(ctx, f.Team)
			if err != nil {
				return fmt.Errorf("fetching config: %w", err)
			}

			if len(configs) == 0 {
				out.Infoln("No config found")
				return nil
			}

			var summaries []ConfigSummary
			for _, c := range configs {
				envName := c.TeamEnvironment.Environment.Name

				if len(f.Environment) > 0 && !slices.Contains(f.Environment, envName) {
					continue
				}

				var keyNames []string
				for _, v := range c.Values {
					keyNames = append(keyNames, v.Name)
				}

				var workloadNames []string
				for _, w := range c.Workloads.Nodes {
					workloadNames = append(workloadNames, w.GetName())
				}

				summaries = append(summaries, ConfigSummary{
					Name:         c.Name,
					Environment:  envName,
					Keys:         summarizeList(keyNames),
					Workloads:    summarizeList(workloadNames),
					LastModified: config.LastModified(c.LastModifiedAt),
				})
			}

			if len(summaries) == 0 {
				out.Infoln("No config matches the given filters")
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
// e.g. ["a", "b", "c", "d", "e"] -> "a, b, c, +2 more"
func summarizeList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	if len(items) <= maxListItems {
		return strings.Join(items, ", ")
	}

	return fmt.Sprintf("%s, +%d more", strings.Join(items[:maxListItems], ", "), len(items)-maxListItems)
}
