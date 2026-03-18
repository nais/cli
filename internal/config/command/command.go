package command

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Configs(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Config{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "configs",
		Title:       "Manage configs for a team.",
		StickyFlags: f,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(f.Team)
		},
		SubCommands: []*naistrix.Command{
			list(f),
			activity(f),
			get(f),
			create(f),
			deleteConfig(f),
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

func metadataFromArgs(args *naistrix.Arguments, team string, environment string) config.Metadata {
	return config.Metadata{
		TeamSlug:        team,
		EnvironmentName: environment,
		Name:            args.Get("name"),
	}
}

func autoCompleteConfigNames(ctx context.Context, team, environment string, requireEnvironment bool) ([]string, string) {
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
	return autoCompleteConfigNamesInEnvironments(ctx, team, environments, requireEnvironment)
}

func autoCompleteConfigNamesInEnvironments(ctx context.Context, team string, environments []string, requireEnvironment bool) ([]string, string) {
	if team == "" {
		return nil, "Please provide team to auto-complete config names. 'nais config set team <team>', or '--team <team>' flag."
	}
	if requireEnvironment && len(environments) == 0 {
		return nil, "Please provide environment to auto-complete config names. '--environment <environment>' flag."
	}

	environmentFilter := make(map[string]struct{}, len(environments))
	for _, env := range environments {
		if env == "" {
			continue
		}
		environmentFilter[env] = struct{}{}
	}

	configs, err := config.GetAll(ctx, team)
	if err != nil {
		return nil, fmt.Sprintf("Unable to fetch configs for auto-completion: %v", err)
	}

	seen := make(map[string]struct{})
	var names []string
	for _, c := range configs {
		if len(environmentFilter) > 0 {
			if _, ok := environmentFilter[c.TeamEnvironment.Environment.Name]; !ok {
				continue
			}
		}
		if _, ok := seen[c.Name]; ok {
			continue
		}
		seen[c.Name] = struct{}{}
		names = append(names, c.Name)
	}
	sort.Strings(names)

	if len(names) == 0 && len(environmentFilter) > 0 {
		sortedEnvironments := make([]string, 0, len(environmentFilter))
		for env := range environmentFilter {
			sortedEnvironments = append(sortedEnvironments, env)
		}
		sort.Strings(sortedEnvironments)
		if len(sortedEnvironments) == 1 {
			return nil, fmt.Sprintf("No configs found in environment %q.", sortedEnvironments[0])
		}
		return nil, fmt.Sprintf("No configs found in environments: %s.", strings.Join(sortedEnvironments, ", "))
	}

	return names, "Select a config."
}
