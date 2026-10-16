package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Secrets(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Secret{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:              "secret",
		Aliases:           []string{"secrets"},
		Title:             "Manage secrets for a team.",
		Description:       "Commands for listing, creating, viewing, updating, and deleting secrets for a team across environments.",
		StickyFlags:       f,
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
		SubCommands: []*naistrix.Command{
			list(f),
			activity(f),
			get(f),
			create(f),
			deleteSecret(f),
			set(f),
			unset(f),
		},
	}
}

var defaultArgs = []naistrix.Argument{
	{Name: "name", Prompt: "Name of the secret"},
}

func validateArgs(_ context.Context, args *naistrix.Arguments) error {
	if args.Len() != 1 {
		return fmt.Errorf("expected 1 argument, got %d", args.Len())
	}
	if args.Get("name") == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func metadataFromArgs(args *naistrix.Arguments, team string, environment string) secret.Metadata {
	return secret.Metadata{
		TeamSlug:        team,
		EnvironmentName: environment,
		Name:            args.Get("name"),
	}
}

func secretNameChoices(flags *flag.Secret) naistrix.ArgumentChoicesFunc {
	return func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
		secrets, err := secret.GetAll(ctx, flags.Team, gql.SecretFilter{
			Environments: []string{string(flags.Environment)},
		})
		if err != nil {
			return nil, err
		}

		names := make([]string, len(secrets))
		for i, s := range secrets {
			names[i] = s.Name
		}
		slices.Sort(names)
		return names, nil
	}
}
