package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func listValkeys(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.List{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List existing Valkey instances.",
		Description: "This command lists all Valkey instances for a given team.",
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
				Description: "List all Valkeys for the team named my-team.",
				Command:     "my-team",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			valkeys, err := valkey.GetAll(ctx, args[0])
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			// TODO: flags to filter by environment, memory, tier, etc?
			data := pterm.TableData{
				{
					"Environment",
					"Name",
					"Tier",
					"Memory",
					"Workloads",
					"Max memory policy",
					"State",
				},
			}
			for _, v := range valkeys {
				data = append(data, []string{
					v.TeamEnvironment.Environment.Name,
					v.Name,
					string(v.Tier),
					string(v.Memory),
					fmt.Sprintf("%d", len(v.Access.Edges)),
					string(v.MaxMemoryPolicy),
					string(v.State),
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(data).Render()
		},
	}
}
