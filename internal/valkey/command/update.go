package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

func updateValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Update{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "update",
		Title:       "Update a Valkey instance.",
		Description: "This command updates an existing Valkey instance.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}
			if err := flags.Validate(); err != nil {
				return err
			}

			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteValkeyNames(ctx, flags.Team, string(flags.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Set the |MEMORY| for a Valkey instance named some-valkey.",
				Command:     "some-valkey --memory GB_8",
			},
			{
				Description: "Set the |TIER| for a Valkey instance named some-valkey.",
				Command:     "some-valkey --tier SINGLE_NODE",
			},
			{
				Description: "Set the |MAX_MEMORY_POLICY| for a Valkey instance named some-valkey.",
				Command:     "some-valkey --max-memory-policy NO_EVICTION",
			},
			{
				Description: "Set all available options for a Valkey instance named some-valkey.",
				Command:     "some-valkey --memory GB_8 --tier SINGLE_NODE --max-memory-policy NO_EVICTION",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			data := &valkey.Valkey{
				Tier:            existing.Tier,
				Memory:          existing.Memory,
				MaxMemoryPolicy: existing.MaxMemoryPolicy,
			}

			info := [][]string{
				{"Field", "Old Value", "New Value"},
				{"Team", metadata.TeamSlug, "(unchanged)"},
				{"Environment", metadata.EnvironmentName, "(unchanged)"},
				{"Name", metadata.Name, "(unchanged)"},
			}

			newTier := "(unchanged)"
			if flags.Tier != "" && string(flags.Tier) != string(existing.Tier) {
				data.Tier = gql.ValkeyTier(flags.Tier)
				if flags.IsVerbose() {
					out.Infof("Changing tier from %q to %q\n", existing.Tier, data.Tier)
				}
				newTier = string(data.Tier)
			}
			info = append(info, []string{"Tier", string(existing.Tier), newTier})

			newMemory := "(unchanged)"
			if flags.Memory != "" && string(flags.Memory) != string(existing.Memory) {
				data.Memory = gql.ValkeyMemory(flags.Memory)
				if flags.IsVerbose() {
					out.Infof("Changing memory from %q to %q\n", existing.Memory, data.Memory)
				}
				newMemory = string(data.Memory)
			}
			info = append(info, []string{"Memory", string(existing.Memory), newMemory})

			newMaxMemoryPolicy := "(unchanged)"
			if flags.MaxMemoryPolicy != "" && string(flags.MaxMemoryPolicy) != string(existing.MaxMemoryPolicy) {
				data.MaxMemoryPolicy = gql.ValkeyMaxMemoryPolicy(flags.MaxMemoryPolicy)
				if flags.IsVerbose() {
					out.Infof("Changing max memory policy from %q to %q\n", existing.MaxMemoryPolicy, data.MaxMemoryPolicy)
				}
				newMaxMemoryPolicy = string(data.MaxMemoryPolicy)
			}
			info = append(info, []string{"Max memory policy", string(existing.MaxMemoryPolicy), newMaxMemoryPolicy})

			out.Infoln("You are about to update a Valkey instance with the following configuration:")
			if err := out.Table(output.TableWithMargins()).Render(info); err != nil {
				return err
			}

			if !flags.Yes {
				out.Warnln("Changing settings may cause a restart of the Valkey instance.")
				if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
					return err
				} else if !result {
					return fmt.Errorf("cancelled by user")
				}
			}

			if _, err = valkey.Update(ctx, metadata, data); err != nil {
				return err
			}

			out.Successf("Updated Valkey instance %q for %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
