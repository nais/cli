package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func create(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Create{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create a new secret.",
		Description: "This command creates a new empty secret in a team environment.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			return validateArgs(args)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create a secret named my-secret in environment dev.",
				Command:     "my-secret --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			_, err := secret.Create(ctx, metadata)
			if err != nil {
				return fmt.Errorf("creating secret: %w", err)
			}

			out.Successf("Created secret %q in %q for team %q\n", metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
			return nil
		},
	}
}
