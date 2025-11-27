package command

import (
	"context"
	"fmt"
	"slices"

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

			// TODO: remove pterm here and for rest of commands that use it
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
			for _, o := range opensearches {
				// TODO: use filter in GQL query instead
				if len(flags.Environment) > 0 && !slices.Contains(flags.Environment, string(o.TeamEnvironment.Environment.Name)) {
					continue
				}
				data = append(data, []string{
					o.TeamEnvironment.Environment.Name,
					o.Name,
					string(o.Tier),
					string(o.Memory),
					fmt.Sprintf("%d GB", o.StorageGB),
					fmt.Sprintf("%d", len(o.Access.Edges)),
					o.Version.Actual,
					string(o.State),
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(data).Render()
		},
	}
}
