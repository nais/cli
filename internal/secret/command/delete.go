package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
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

			out.Warnln("You are about to delete a secret with the following configuration:")
			if err := out.Table().Render(secret.FormatDetails(metadata, existing)); err != nil {
				return err
			}

			if len(existing.Workloads.Nodes) > 0 {
				out.Warnln("This secret is currently in use by the following workloads:")
				if err := out.Table().Render(secret.FormatWorkloads(existing)); err != nil {
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

			if deleted, err := secret.Delete(ctx, metadata); err != nil {
				return fmt.Errorf("deleting secret: %w", err)
			} else if deleted {
				out.Successf("Deleted secret %q from %q for team %q\n", metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
			}

			return nil
		},
	}
}
