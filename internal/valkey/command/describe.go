package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func describeValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Describe{Valkey: parentFlags}
	return &naistrix.Command{
		Name:         "describe",
		Title:        "Describe a Valkey instance.",
		Description:  "This command describes a Valkey instance.",
		Flags:        flags,
		Args:         args,
		ValidateFunc: validateFunc,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			pterm.DefaultSection.Println("Valkey instance details")
			err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(pterm.TableData{
				{"Team", "Environment", "Name", "Size", "Tier", "Max Memory Policy"},
				{metadata.TeamSlug, metadata.EnvironmentName, metadata.Name, string(existing.Size), string(existing.Tier), string(existing.MaxMemoryPolicy)},
			}).Render()
			if err != nil {
				return fmt.Errorf("rendering table: %w", err)
			}

			pterm.DefaultSection.Println("Valkey access list")
			acl := pterm.TableData{
				{"Team", "Environment", "Name", "Type", "Access"},
			}
			for _, edge := range existing.Access.Edges {
				acl = append(acl, []string{
					edge.Node.Workload.GetTeam().Slug,
					metadata.EnvironmentName,
					edge.Node.Workload.GetName(),
					edge.Node.Workload.GetTypename(),
					edge.Node.Access,
				})
			}
			err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(acl).Render()
			return nil
		},
	}
}
