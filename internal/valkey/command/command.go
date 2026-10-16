package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func Valkey(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Valkey{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:              "valkey",
		Aliases:           []string{"valkeys"},
		Title:             "Manage Valkey instances.",
		Description:       "Commands for creating, updating, deleting, and inspecting Valkey instances and their credentials.",
		StickyFlags:       f,
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
		SubCommands: []*naistrix.Command{
			create(f),
			credentials(f),
			delete(f),
			get(f),
			list(f),
			updateValkey(f),
		},
	}
}

var defaultArgs = []naistrix.Argument{
	{Name: "name", Prompt: "Name of the Valkey instance"},
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

func metadataFromArgs(args *naistrix.Arguments, team string, environment string) valkey.Metadata {
	return valkey.Metadata{
		TeamSlug:        team,
		EnvironmentName: environment,
		Name:            args.Get("name"),
	}
}

func valkeyNameChoices(flags *flag.Valkey) naistrix.ArgumentChoicesFunc {
	return func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
		instances, err := valkey.GetAll(ctx, flags.Team, gql.ValkeyFilter{
			Environments: []string{string(flags.Environment)},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to fetch Valkey instances: %w", err)
		}

		names := make([]string, len(instances))
		for i, instance := range instances {
			names[i] = instance.Name
		}
		slices.Sort(names)
		return names, nil
	}
}
