package command

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Config(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Config{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:         "config",
		Title:        "Manage config for a team.",
		Description:  "Commands for listing, creating, viewing, updating, and deleting configuration values for a team across environments.",
		StickyFlags:  f,
		ValidateFunc: validation.RequireTeam(f),
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

func validateArgs(_ context.Context, args *naistrix.Arguments) error {
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

func autoCompleteConfigNames(flags *flag.Config) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != 0 {
			return nil, ""
		}

		if flags.Team == "" {
			return nil, "Please provide team to auto-complete config names. 'nais defaults set team <team>', or '--team <team>' flag."
		}

		if flags.Environment == "" {
			return nil, "Please provide environment to auto-complete config names. 'nais defaults set environment <env>', or '--environment <env>' flag."
		}

		configs, err := config.GetAll(ctx, flags.Team, gql.ConfigFilter{
			Environments: []string{string(flags.Environment)},
		})
		if err != nil {
			return nil, fmt.Sprintf("Unable to fetch config for auto-completion: %v", err)
		}

		seen := make(map[string]struct{})
		var names []string
		for _, c := range configs {
			if string(flags.Environment) != c.TeamEnvironment.Environment.Name {
				continue
			}

			if _, ok := seen[c.Name]; ok {
				continue
			}
			seen[c.Name] = struct{}{}
			names = append(names, c.Name)
		}
		sort.Strings(names)

		if len(names) == 0 {
			return nil, "No config found."
		}

		return names, "Select a config."
	}
}
