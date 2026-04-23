package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func deleteConfig(parentFlags *flag.Config) *naistrix.Command {
	f := &flag.Delete{Config: parentFlags}
	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete a config.",
		Description: "This command deletes a config and all its values.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteConfigNames(ctx, f.Team, string(f.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Delete a config named my-config in environment dev.",
				Command:     "my-config --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			existing, err := config.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching config: %w", err)
			}

			out.Warnln("You are about to delete a config with the following configuration:")
			if err := out.Table().Render(config.FormatDetails(metadata, existing)); err != nil {
				return err
			}

			if len(existing.Workloads.Nodes) > 0 {
				out.Warnln("This config is currently in use by the following workloads:")
				if err := out.Table().Render(config.FormatWorkloads(existing)); err != nil {
					return err
				}
			}

			if !f.Yes {
				if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
					return err
				} else if !result {
					return fmt.Errorf("cancelled by user")
				}
			}

			deleted, err := config.Delete(ctx, metadata)
			if err != nil {
				return fmt.Errorf("deleting config: %w", err)
			}

			if !deleted {
				out.Warnf("Config %q in %q for team %q was not deleted\n", metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
				return nil
			}

			out.Successf("Deleted config %q from %q for team %q\n", metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
			return nil
		},
	}
}
