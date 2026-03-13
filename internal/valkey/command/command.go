package command

import (
	"context"
	"fmt"
	"sort"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func Valkey(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.Valkey{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "valkey",
		Aliases:     []string{"valkeys"},
		Title:       "Manage Valkey instances.",
		StickyFlags: flags,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		SubCommands: []*naistrix.Command{
			create(flags),
			delete(flags),
			get(flags),
			list(flags),
			updateValkey(flags),
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
		return nil, "Please provide team to auto-complete Valkey instance names. 'nais config set team <team>', or '--team <team>' flag."
	}
	if requireEnvironment && environment == "" {
		return nil, "Please provide environment to auto-complete Valkey instance names. '--environment <environment>' flag."
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
