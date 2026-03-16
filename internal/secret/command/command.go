package command

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

func autoCompleteSecretNames(ctx context.Context, team, environment string, requireEnvironment bool) ([]string, string) {
	if countEnvironmentFlagsInCLIArgs() > 1 {
		return nil, "Only one --environment/-e flag may be provided."
	}

	if environment == "" {
		envs := environmentValuesFromCLIArgs()
		if len(envs) == 1 {
			environment = envs[0]
		}
	}

	environments := []string{}
	if environment != "" {
		environments = append(environments, environment)
	}
	return autoCompleteSecretNamesInEnvironments(ctx, team, environments, requireEnvironment)
}

func autoCompleteSecretNamesInEnvironments(ctx context.Context, team string, environments []string, requireEnvironment bool) ([]string, string) {
	if team == "" {
		return nil, "Please provide team to auto-complete secret names. 'nais config set team <team>', or '--team <team>' flag."
	}
	if requireEnvironment && len(environments) == 0 {
		return nil, "Please provide environment to auto-complete secret names. '--environment <environment>' flag."
	}

	environmentFilter := make(map[string]struct{}, len(environments))
	for _, env := range environments {
		if env == "" {
			continue
		}
		environmentFilter[env] = struct{}{}
	}

	secrets, err := secret.GetAll(ctx, team)
	if err != nil {
		return nil, fmt.Sprintf("Unable to fetch secrets for auto-completion: %v", err)
	}

	seen := make(map[string]struct{})
	var names []string
	for _, s := range secrets {
		if len(environmentFilter) > 0 {
			if _, ok := environmentFilter[s.TeamEnvironment.Environment.Name]; !ok {
				continue
			}
		}
		if _, ok := seen[s.Name]; ok {
			continue
		}
		seen[s.Name] = struct{}{}
		names = append(names, s.Name)
	}
	sort.Strings(names)

	if len(names) == 0 && len(environmentFilter) > 0 {
		sortedEnvironments := make([]string, 0, len(environmentFilter))
		for env := range environmentFilter {
			sortedEnvironments = append(sortedEnvironments, env)
		}
		sort.Strings(sortedEnvironments)
		if len(sortedEnvironments) == 1 {
			return nil, fmt.Sprintf("No secrets found in environment %q.", sortedEnvironments[0])
		}
		return nil, fmt.Sprintf("No secrets found in environments: %s.", strings.Join(sortedEnvironments, ", "))
	}

	return names, "Select a secret."
}
