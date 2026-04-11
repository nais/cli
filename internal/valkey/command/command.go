package command

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func Valkey(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Valkey{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "valkey",
		Aliases:     []string{"valkeys"},
		Title:       "Manage Valkey instances.",
		Description: "Commands for creating, updating, deleting, and inspecting Valkey instances and their credentials.",
		StickyFlags: f,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(f.Team)
		},
		SubCommands: []*naistrix.Command{
			create(f),
			credentials(f),
			delete(f),
			get(f),
			list(f),
			proxy(f),
			updateValkey(f),
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

func metadataFromArgs(args *naistrix.Arguments, team string, environment string) valkey.Metadata {
	return valkey.Metadata{
		TeamSlug:        team,
		EnvironmentName: environment,
		Name:            args.Get("name"),
	}
}

func autoCompleteValkeyNames(ctx context.Context, team, environment string, requireEnvironment bool) ([]string, string) {
	if team == "" {
		return nil, "Please provide team to auto-complete Valkey instance names. 'nais defaults set team <team>', or '--team <team>' flag."
	}

	if environmentFlagOccurrencesFromCLIArgs() > 1 {
		return nil, "Please specify exactly one environment to auto-complete Valkey instance names. '-e, --environment <environment>' flag."
	}

	if environment == "" {
		envs := environmentValuesFromCLIArgs()
		if len(envs) == 1 {
			environment = envs[0]
		}
	}

	if requireEnvironment && environment == "" {
		return nil, "Please provide environment to auto-complete Valkey instance names. '-e, --environment <environment>' flag."
	}

	instances, err := valkey.GetAll(ctx, team)
	if err != nil {
		return nil, "Unable to fetch Valkey instances."
	}

	seen := make(map[string]struct{})
	var names []string
	for _, instance := range instances {
		if environment != "" && instance.TeamEnvironment.Environment.Name != environment {
			continue
		}
		if _, ok := seen[instance.Name]; ok {
			continue
		}
		seen[instance.Name] = struct{}{}
		names = append(names, instance.Name)
	}

	sort.Strings(names)
	if len(names) == 0 && environment != "" {
		return nil, fmt.Sprintf("No Valkey instances found in environment %q.", environment)
	}

	return names, "Select a Valkey instance."
}

func environmentValuesFromCLIArgs() []string {
	return cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
}

func environmentFlagOccurrencesFromCLIArgs() int {
	return cliflags.CountFlagOccurrences(os.Args, "-e", "--environment")
}

func validateSingleEnvironmentFlagUsage() error {
	if environmentFlagOccurrencesFromCLIArgs() > 1 {
		return fmt.Errorf("only one -e, --environment flag may be provided")
	}
	return nil
}
