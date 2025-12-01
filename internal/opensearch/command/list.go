package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

type state string

func (s state) String() string {
	switch s {
	case state(gql.OpenSearchStateRunning):
		return "Running"
	case state(gql.OpenSearchStatePoweroff):
		return "<error>Stopped</error>"
	case state(gql.OpenSearchStateRebalancing):
		return "<warn>Rebalancing</warn>"
	case state(gql.OpenSearchStateRebuilding):
		return "<info>Rebuilding</info>"
	default:
		return "<info>Unknown</info>"
	}
}

type OpenSearchSummary struct {
	State       state  `header:"State"`
	Environment string `header:"Environment"`
	Name        string `header:"Name"`
	Tier        string `header:"Tier"`
	Memory      string `header:"Memory"`
	StorageGB   int    `header:"Storage (GB)"`
	Workloads   int    `header:"Workloads"`
	Version     string `header:"Version"`
}

func list(parentFlags *flag.OpenSearch) *naistrix.Command {
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

			if len(opensearches) == 0 {
				out.Infoln("No OpenSearch instances found")
				return nil
			}

			var summaries []OpenSearchSummary
			for _, o := range opensearches {
				// TODO: use filter in GQL query instead
				if len(flags.Environment) > 0 && !slices.Contains(flags.Environment, string(o.TeamEnvironment.Environment.Name)) {
					continue
				}
				summaries = append(summaries, OpenSearchSummary{
					Environment: o.TeamEnvironment.Environment.Name,
					Name:        o.Name,
					Tier:        string(o.Tier),
					Memory:      string(o.Memory),
					StorageGB:   o.StorageGB,
					Workloads:   len(o.Access.Edges),
					Version:     o.Version.Actual,
					State:       state(o.State),
				})
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(summaries)
			}

			return out.Table().Render(summaries)
		},
	}
}
