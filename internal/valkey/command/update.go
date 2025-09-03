package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func updateValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Update{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "update",
		Title:       "Update a Valkey instance.",
		Description: "This command updates an existing Valkey instance.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args []string) error {
			if err := flags.Validate(); err != nil {
				return err
			}
			return defaultValidateFunc(ctx, args)
		},
		Examples: []naistrix.Example{
			{
				Description: "Set the |SIZE| for a Valkey instance named some-valkey for my-team in the dev environment.",
				Command:     "my-team dev some-valkey --size RAM_8GB",
			},
			{
				Description: "Set the |TIER| for a Valkey instance named some-valkey for my-team in the dev environment.",
				Command:     "my-team dev some-valkey --tier SINGLE_NODE",
			},
			{
				Description: "Set the |MAX_MEMORY_POLICY| for a Valkey instance named some-valkey for my-team in the dev environment.",
				Command:     "my-team dev some-valkey --max-memory-policy NO_EVICTION",
			},
			{
				Description: "Set all available options for a Valkey instance named some-valkey for my-team in the dev environment.",
				Command:     "my-team dev some-valkey --size RAM_8GB --tier SINGLE_NODE --max-memory-policy NO_EVICTION",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			data := &valkey.Valkey{
				Size:            existing.Size,
				Tier:            existing.Tier,
				MaxMemoryPolicy: existing.MaxMemoryPolicy,
			}

			info := pterm.TableData{
				{"Field", "Old Value", "New Value"},
				{"Team", metadata.TeamSlug, "(unchanged)"},
				{"Environment", metadata.EnvironmentName, "(unchanged)"},
				{"Name", metadata.Name, "(unchanged)"},
			}

			newSize := "(unchanged)"
			if flags.Size != "" && string(flags.Size) != string(existing.Size) {
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing size from %q to %q\n", existing.Size, flags.Size)
				}
				data.Size = gql.ValkeySize(flags.Size)
				newSize = string(flags.Size)
			}
			info = append(info, []string{"Size", string(existing.Size), newSize})

			newTier := "(unchanged)"
			if flags.Tier != "" && string(flags.Tier) != string(existing.Tier) {
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing tier from %q to %q\n", existing.Tier, flags.Tier)
				}
				data.Tier = gql.ValkeyTier(flags.Tier)
				newTier = string(flags.Tier)
			}
			info = append(info, []string{"Tier", string(existing.Tier), newTier})

			newMaxMemoryPolicy := "(unchanged)"
			if flags.MaxMemoryPolicy != "" && string(flags.MaxMemoryPolicy) != string(existing.MaxMemoryPolicy) {
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing max memory policy from %q to %q\n", existing.MaxMemoryPolicy, flags.MaxMemoryPolicy)
				}
				data.MaxMemoryPolicy = gql.ValkeyMaxMemoryPolicy(flags.MaxMemoryPolicy)
				newMaxMemoryPolicy = string(flags.MaxMemoryPolicy)
			}
			info = append(info, []string{"Max memory policy", string(existing.MaxMemoryPolicy), newMaxMemoryPolicy})

			pterm.Info.Println("You are about to update a Valkey instance with the following configuration:")
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(info).Render(); err != nil {
				return err
			}

			pterm.Warning.Println("Changing settings may cause a restart of the Valkey instance.")
			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			_, err = valkey.Update(ctx, metadata, data)
			if err != nil {
				return err
			}

			pterm.Success.Printf("Updated Valkey instance %q for %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
