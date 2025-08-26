package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func deleteValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Delete{Valkey: parentFlags}
	return &naistrix.Command{
		Name:         "delete",
		Title:        "Delete a Valkey instance.",
		Description:  "This command deletes an existing Valkey instance.",
		Flags:        flags,
		Args:         args,
		ValidateFunc: validateFunc,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			pterm.Warning.Println("You are about to delete a Valkey instance with the following configuration:")
			outData := pterm.TableData{
				{"Field", "Value"},
				{"Team", metadata.TeamSlug},
				{"Environment", metadata.EnvironmentName},
				{"Name", metadata.Name},
				{"Size", string(existing.Size)},
				{"Tier", string(existing.Tier)},
				{"Max Memory Policy", string(existing.MaxMemoryPolicy)},
			}
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(outData).Render(); err != nil {
				return err
			}
			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			deleted, err := valkey.Delete(ctx, metadata)
			if err != nil {
				return err
			}

			if deleted {
				pterm.Success.Printf("Deleted Valkey instance %q from %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			}
			return nil
		},
	}
}
