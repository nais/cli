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

func unset(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Unset{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "unset",
		Title:       "Unset a key from a secret.",
		Description: "This command removes a key-value pair from a secret.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			if err := validateArgs(args); err != nil {
				return err
			}
			if f.Key == "" {
				return fmt.Errorf("--key is required")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteSecretNames(ctx, f.Team, string(f.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Unset a key from a secret.",
				Command:     "my-secret --environment dev --key OLD_API_KEY",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			out.Warnf("You are about to unset key %q from secret %q in %q.\n", f.Key, metadata.Name, metadata.EnvironmentName)
			if !f.Yes {
				if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
					return err
				} else if !result {
					return fmt.Errorf("cancelled by user")
				}
			}

			if err := secret.RemoveValue(ctx, metadata, f.Key); err != nil {
				return fmt.Errorf("unsetting secret key: %w", err)
			}

			out.Successf("Unset key %q from secret %q in %q\n", f.Key, metadata.Name, metadata.EnvironmentName)
			return nil
		},
	}
}
