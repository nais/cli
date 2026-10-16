package command

import (
	"context"
	"fmt"
	"slices"

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
		Name:              "config",
		Title:             "Manage config for a team.",
		Description:       "Commands for listing, creating, viewing, updating, and deleting configuration values for a team across environments.",
		StickyFlags:       f,
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
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
	{Name: "name", Prompt: "Name of the config"},
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

func configNameChoices(flags *flag.Config) naistrix.ArgumentChoicesFunc {
	return func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
		configs, err := config.GetAll(ctx, flags.Team, gql.ConfigFilter{
			Environments: []string{string(flags.Environment)},
		})
		if err != nil {
			return nil, err
		}

		names := make([]string, len(configs))
		for i, c := range configs {
			names[i] = c.Name
		}
		slices.Sort(names)
		return names, nil
	}
}
