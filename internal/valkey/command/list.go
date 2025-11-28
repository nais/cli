package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

type state string

func (s state) String() string {
	switch s {
	case state(gql.ValkeyStateRunning):
		return "Running"
	case state(gql.ValkeyStatePoweroff):
		return "<error>Stopped</error>"
	case state(gql.ValkeyStateRebalancing):
		return "<warn>Rebalancing</warn>"
	case state(gql.ValkeyStateRebuilding):
		return "<info>Rebuilding</info>"
	default:
		return "<info>Unknown</info>"
	}
}

type ValkeySummary struct {
	State       state  `header:"State"`
	Environment string `header:"Environment"`
	Name        string `header:"Name"`
	Tier        string `header:"Tier"`
	Memory      string `header:"Memory"`
	Workloads   int    `header:"Workloads"`
}

func listValkeys(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.List{Valkey: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List existing Valkey instances.",
		Description: "This command lists all Valkey instances for a given team.",
		Flags:       flags,
		Examples: []naistrix.Example{
			{
				Description: "List all Valkeys for the team.",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			valkeys, err := valkey.GetAll(ctx, flags.Team)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			if len(valkeys) == 0 {
				out.Infoln("No Valkey instances found")
				return nil
			}

			var summaries []ValkeySummary
			for _, v := range valkeys {
				// TODO: use filter in GQL query instead
				if len(flags.Environment) > 0 && !slices.Contains(flags.Environment, string(v.TeamEnvironment.Environment.Name)) {
					continue
				}
				summaries = append(summaries, ValkeySummary{
					Environment: v.TeamEnvironment.Environment.Name,
					Name:        v.Name,
					Tier:        string(v.Tier),
					Memory:      string(v.Memory),
					Workloads:   len(v.Access.Edges),
					State:       state(v.State),
				})
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(summaries)
			}

			return out.Table().Render(summaries)
		},
	}
}
