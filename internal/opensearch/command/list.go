package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
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
		Args: []naistrix.Argument{
			{Name: "team"},
		},
		ValidateFunc: func(_ context.Context, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("team cannot be empty")
			}
			return nil
		},
		Examples: []naistrix.Example{
			{
				Description: "List all OpenSearches for the team named my-team.",
				Command:     "my-team",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			opensearches, err := opensearch.GetAll(ctx, args[0])
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
