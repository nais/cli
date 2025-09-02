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
	flags := &flag.Create{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create a Valkey instance.",
		Description: "This command creates a Valkey instance.",
		Flags:       flags,
		Args:        defaultArgs,
		Examples: []naistrix.Example{
			{
				Description: "Create a Valkey instance named some-valkey for my-team in the dev environment, using the default size and tier.",
				Command:     "my-team dev some-valkey",
			},
			{
				Description: "Create a Valkey instance named some-valkey for my-team in the dev environment, with the specified |SIZE|.",
				Command:     "my-team dev some-valkey --size RAM_4GB",
			},
			{
				Description: "Create a Valkey instance named some-valkey for my-team in the dev environment, with the specified |TIER|.",
				Command:     "my-team dev some-valkey --tier SINGLE_NODE",
			},
			{
				Description: "Create a Valkey instance named some-valkey for my-team in the dev environment, with the specified |MAX_MEMORY_POLICY|.",
				Command:     "my-team dev some-valkey --max-memory-policy ALLKEYS_LRU",
			},
			{
				Description: "Create a Valkey instance named some-valkey for my-team in the dev environment, with all possible options specified.",
				Command:     "my-team dev some-valkey --size RAM_4GB --tier SINGLE_NODE --max-memory-policy ALLKEYS_LRU",
			},
		},
		ValidateFunc: defaultValidateFunc,
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
	}
}
