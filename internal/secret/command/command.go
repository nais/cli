package command

import (
	"context"
	"fmt"
	"sort"

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
		Name:         "secret",
		Aliases:      []string{"secrets"},
		Title:        "Manage secrets for a team.",
		Description:  "Commands for listing, creating, viewing, updating, and deleting secrets for a team across environments.",
		StickyFlags:  f,
		ValidateFunc: validation.RequireTeam(f),
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
	{Name: "name"},
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

func autoCompleteSecretNames(flags *flag.Secret) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() > 0 {
			return nil, ""
		}

		if flags.Team == "" {
			return nil, "Please provide team to auto-complete secret names. 'nais defaults set team <team>', or '--team <team>' flag."
		}

		if flags.Environment == "" {
			return nil, "Please provide environment to auto-complete secret names. '-e, --environment <environment>' flag."
		}

		secrets, err := secret.GetAll(ctx, flags.Team, gql.SecretFilter{
			Environments: []string{string(flags.Environment)},
		})
		if err != nil {
			return nil, fmt.Sprintf("Unable to fetch secrets for auto-completion: %v", err)
		}

		seen := make(map[string]struct{})
		var names []string
		for _, s := range secrets {
			if string(flags.Environment) != s.TeamEnvironment.Environment.Name {
				continue
			}
			if _, ok := seen[s.Name]; ok {
				continue
			}
			seen[s.Name] = struct{}{}
			names = append(names, s.Name)
		}
		sort.Strings(names)

		if len(names) == 0 {
			return nil, "No secrets found."
		}

		return names, "Select a secret."
	}
}
