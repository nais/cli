package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func deleteSecret(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Delete{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete a secret.",
		Description: "This command deletes a secret and all its values.",
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
				return autoCompleteSecretNames(ctx, f.Team, string(f.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Delete a secret named my-secret in environment dev.",
				Command:     "my-secret --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			existing, err := secret.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching secret: %w", err)
			}

			pterm.Warning.Println("You are about to delete a secret with the following configuration:")
			err = pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(secret.FormatDetails(metadata, existing)).
				Render()
			if err != nil {
				return err
			}

			if len(existing.Workloads.Nodes) > 0 {
				pterm.Warning.Println("This secret is currently in use by the following workloads:")
				err = pterm.DefaultTable.
					WithHasHeader().
					WithHeaderRowSeparator("-").
					WithData(secret.FormatWorkloads(existing)).
					Render()
				if err != nil {
					return err
				}
			}

			if !f.Yes {
				result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
				if !result {
					return fmt.Errorf("cancelled by user")
				}
			}

			deleted, err := secret.Delete(ctx, metadata)
			if err != nil {
				return fmt.Errorf("deleting secret: %w", err)
			}

			if deleted {
				pterm.Success.Printfln("Deleted secret %q from %q for team %q", metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
			}

			return nil
		},
	}
}
