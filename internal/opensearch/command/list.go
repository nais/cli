package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func listOpenSearches(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.List{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List existing Opensearch instances.",
		Description: "This command lists all Opensearch instances for a given team.",
		Flags:       flags,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		Examples: []naistrix.Example{
			{
				Description: "List all OpenSearches for the team.",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			opensearches, err := opensearch.GetAll(ctx, flags.Team)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			// TODO: flags to filter by environment, memory, tier, etc?
			data := pterm.TableData{
				{
					"Environment",
					"Name",
					"Tier",
					"Memory",
					"Storage",
					"Workloads",
					"Version",
					"State",
				},
			}
			for _, v := range opensearches {
				data = append(data, []string{
					v.TeamEnvironment.Environment.Name,
					v.Name,
					string(v.Tier),
					string(v.Memory),
					fmt.Sprintf("%d GB", v.StorageGB),
					fmt.Sprintf("%d", len(v.Access.Edges)),
					v.Version.Actual,
					string(v.State),
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(data).Render()
		},
	}
}
