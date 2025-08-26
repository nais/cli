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

func createValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Upsert{Valkey: parentFlags}
	return &naistrix.Command{
		Name:         "create",
		Title:        "Create a Valkey instance.",
		Description:  "This command creates a Valkey instance.",
		Flags:        flags,
		Args:         args,
		ValidateFunc: validateFunc,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			// defaults
			data := &valkey.Valkey{
				Size: "RAM_1GB",
				Tier: "HIGH_AVAILABILITY",
			}

			if flags.Size != "" {
				data.Size = gql.ValkeySize(flags.Size)
			}
			if flags.Tier != "" {
				data.Tier = gql.ValkeyTier(flags.Tier)
			}

			outData := pterm.TableData{
				{"Field", "Value"},
				{"Team", metadata.TeamSlug},
				{"Environment", metadata.EnvironmentName},
				{"Name", metadata.Name},
				{"Size", string(data.Size)},
				{"Tier", string(data.Tier)},
			}

			if flags.MaxMemoryPolicy != "" {
				data.MaxMemoryPolicy = gql.ValkeyMaxMemoryPolicy(flags.MaxMemoryPolicy)
				outData = append(outData, []string{"Max Memory Policy", string(data.MaxMemoryPolicy)})
			}

			pterm.Info.Println("You are about to create a Valkey instance with the following configuration:")
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(outData).Render(); err != nil {
				return err
			}
			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			_, err := valkey.Create(ctx, metadata, data)
			if err != nil {
				return err
			}

			pterm.Success.Printfln("Created Valkey instance %q for %q in %q", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
		// TODO: completion, examples, etc.
		//  how do we generate valid options for size and tier in usage text?
		//  how do we display defaults for size and tier?
		//  should team and environment be flags? default to some stored state for the current authenticated user?
	}
}
