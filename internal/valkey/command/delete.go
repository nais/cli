package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

func delete(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Delete{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete a Valkey instance.",
		Description: "This command deletes an existing Valkey instance.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
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
				Description: "Delete an existing Valkey instance named some-valkey.",
				Command:     "some-valkey",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			if len(existing.Access.Edges) > 0 {
				out.Errorln("This Valkey instance cannot be deleted as it is currently in use by the following workloads:")
				if err := out.Table(output.TableWithMargins()).Render(valkey.FormatAccessList(metadata, existing)); err != nil {
					return err
				}

				out.Infoln("Remove all references to this Valkey instance from the workloads and try again.")
				return nil
			}

			out.Warnln("You are about to delete a Valkey instance with the following configuration:")
			if err := out.Table(output.TableWithMargins()).Render(valkey.FormatDetails(metadata, existing)); err != nil {
				return err
			}

			if !flags.Yes {
				if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
					return err
				} else if !result {
					return fmt.Errorf("cancelled by user")
				}
			}

			if deleted, err := valkey.Delete(ctx, metadata); err != nil {
				return err
			} else if !deleted {
				return fmt.Errorf("Valkey instance was not deleted")
			}

			out.Successf("Deleted Valkey instance %q from %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
