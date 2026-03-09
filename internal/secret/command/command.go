package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Secrets(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Secret{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "secrets",
		Aliases:     []string{"secret"},
		Title:       "Manage secrets for a team.",
		StickyFlags: f,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(f.Team)
		},
		SubCommands: []*naistrix.Command{
			list(f),
			get(f),
			create(f),
			deleteSecret(f),
			set(f),
			unset(f),
			viewValues(f),
		},
	}
}

var defaultArgs = []naistrix.Argument{
	{Name: "name"},
}

func validateArgs(args *naistrix.Arguments) error {
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

func autoCompleteSecretNames(ctx context.Context, f *flag.Secret) ([]string, string) {
	secrets, err := secret.GetAll(ctx, f.Team)
	if err != nil {
		return nil, "Unable to fetch secrets."
	}
	seen := make(map[string]struct{})
	var names []string
	for _, s := range secrets {
		if _, ok := seen[s.Name]; ok {
			continue
		}
		seen[s.Name] = struct{}{}
		names = append(names, s.Name)
	}
	return names, "Select a secret."
}
