package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func deleteValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Delete{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete a Valkey instance.",
		Description: "This command deletes an existing Valkey instance.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}
			return validateArgs(args)
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
				pterm.Error.Println("This Valkey instance cannot be deleted as it is currently in use by the following workloads:")
				err = pterm.DefaultTable.
					WithHasHeader().
					WithHeaderRowSeparator("-").
					WithData(valkey.FormatAccessList(metadata, existing)).
					Render()
				if err != nil {
					return err
				}

				pterm.Info.Println("Remove all references to this Valkey instance from the workloads and try again.")
				return errors.New("")
			}

			pterm.Warning.Println("You are about to delete a Valkey instance with the following configuration:")
			err = pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(valkey.FormatDetails(metadata, existing)).
				Render()
			if err != nil {
				return err
			}

			if !flags.Yes {
				result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
				if !result {
					return fmt.Errorf("cancelled by user")
				}
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
