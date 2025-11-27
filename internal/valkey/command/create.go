package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
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
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}
			if err := flags.Validate(); err != nil {
				return err
			}

			return validateArgs(args)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create a Valkey instance named some-valkey with default settings.",
				Command:     "some-valkey",
			},
			{
				Description: "Create a Valkey instance named some-valkey with the specified |MEMORY|.",
				Command:     "some-valkey --memory GB_4",
			},
			{
				Description: "Create a Valkey instance named some-valkey with the specified |TIER|.",
				Command:     "some-valkey --tier SINGLE_NODE",
			},
			{
				Description: "Create a Valkey instance named some-valkey with the specified |MAX_MEMORY_POLICY|.",
				Command:     "some-valkey --max-memory-policy ALLKEYS_LRU",
			},
			{
				Description: "Create a Valkey instance named some-valkey with all possible options specified.",
				Command:     "some-valkey --memory GB_4 --tier SINGLE_NODE --max-memory-policy ALLKEYS_LRU",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			// defaults
			data := &valkey.Valkey{
				Memory:          gql.ValkeyMemoryGb1,
				Tier:            gql.ValkeyTierHighAvailability,
				MaxMemoryPolicy: gql.ValkeyMaxMemoryPolicyNoEviction,
			}

			if flags.Memory != "" {
				data.Memory = gql.ValkeyMemory(flags.Memory)
			}
			if flags.Tier != "" {
				data.Tier = gql.ValkeyTier(flags.Tier)
			}
			if flags.MaxMemoryPolicy != "" {
				data.MaxMemoryPolicy = gql.ValkeyMaxMemoryPolicy(flags.MaxMemoryPolicy)
			}

			info := pterm.TableData{
				{"Field", "Value"},
				{"Team", metadata.TeamSlug},
				{"Environment", metadata.EnvironmentName},
				{"Name", metadata.Name},
				{"Tier", string(data.Tier)},
				{"Memory", string(data.Memory)},
				{"Max memory policy", string(data.MaxMemoryPolicy)},
			}

			pterm.Info.Println("You are about to create a Valkey instance with the following configuration:")
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(info).Render(); err != nil {
				return err
			}

			if !flags.Yes {
				result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
				if !result {
					return fmt.Errorf("cancelled by user")
				}
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
