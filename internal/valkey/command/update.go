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
		Name:         "update",
		Title:        "Update a Valkey instance.",
		Description:  "This command updates an existing Valkey instance.",
		Flags:        flags,
		Args:         defaultArgs,
		ValidateFunc: defaultValidateFunc,
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
				Size: existing.Size,
				Tier: existing.Tier,
			}

			outData := pterm.TableData{
				{"Field", "Old Value", "New Value"},
				{"Team", metadata.TeamSlug, "(unchanged)"},
				{"Environment", metadata.EnvironmentName, "(unchanged)"},
				{"Name", metadata.Name, "(unchanged)"},
			}

			if flags.Size != "" {
				data.Size = gql.ValkeySize(flags.Size)
				outData = append(outData, []string{"Size", string(existing.Size), string(data.Size)})
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing size from %q to %q\n", existing.Size, data.Size)
				}
			} else {
				outData = append(outData, []string{"Size", string(existing.Size), "(unchanged)"})
			}

			if flags.Tier != "" {
				data.Tier = gql.ValkeyTier(flags.Tier)
				outData = append(outData, []string{"Tier", string(existing.Tier), string(data.Tier)})
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing tier from %q to %q\n", existing.Tier, data.Tier)
				}
			} else {
				outData = append(outData, []string{"Tier", string(existing.Tier), "(unchanged)"})
			}

			if flags.MaxMemoryPolicy != "" {
				data.MaxMemoryPolicy = gql.ValkeyMaxMemoryPolicy(flags.MaxMemoryPolicy)
				outData = append(outData, []string{"Max Memory Policy", string(existing.MaxMemoryPolicy), string(data.MaxMemoryPolicy)})
				if flags.IsVerbose() {
					pterm.Info.Printf("Updating max memory policy to %q\n", data.MaxMemoryPolicy)
				}
			} else {
				outData = append(outData, []string{"Max Memory Policy", string(existing.MaxMemoryPolicy), "(unchanged)"})
			}

			pterm.Info.Println("You are about to update a Valkey instance with the following configuration:")
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(outData).Render(); err != nil {
				return err
			}
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
