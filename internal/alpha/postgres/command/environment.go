package command

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func postgresNames(instances []postgres.Instance, environment string) []string {
	seen := make(map[string]bool)
	for _, instance := range instances {
		if environment != "" && instance.Environment != environment {
			continue
		}
		name, _, _ := strings.Cut(instance.Name.Name, "/")
		seen[name] = true
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func postgresEnvironments(instances []postgres.Instance, name string) []string {
	seen := make(map[string]bool)
	for _, instance := range instances {
		postgresName, _, _ := strings.Cut(instance.Name.Name, "/")
		if postgresName == name {
			seen[instance.Environment] = true
		}
	}
	envs := make([]string, 0, len(seen))
	for env := range seen {
		envs = append(envs, env)
	}
	sort.Strings(envs)
	return envs
}

func autoCompletePostgresNames(f *flag.Postgres) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != 0 {
			return nil, ""
		}
		if f.Team == "" {
			return nil, "Please provide team to auto-complete Postgres names. 'nais defaults set team <team>', or '--team <team>' flag."
		}
		instances, err := postgres.GetTeamPostgresBranches(ctx, f.Team, nil, nil)
		if err != nil {
			return nil, "Unable to fetch Postgres names."
		}
		names := postgresNames(instances, string(f.Environment))
		if len(names) == 1 {
			return names, ""
		}
		return names, "Select a Postgres instance."
	}
}

func postgresBranches(instances []postgres.Instance, name, environment string) []string {
	seen := make(map[string]bool)
	for _, instance := range instances {
		postgresName, branch, ok := strings.Cut(instance.Name.Name, "/")
		if ok && postgresName == name && instance.Environment == environment {
			seen[branch] = true
		}
	}
	branches := make([]string, 0, len(seen))
	for branch := range seen {
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return branches
}

func autoCompletePostgresBranches(f *flag.Postgres) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() == 0 {
			return autoCompletePostgresNames(f)(ctx, args, "")
		}
		if args.Len() != 1 || f.Team == "" {
			return nil, ""
		}
		instances, err := postgres.GetTeamPostgresBranches(ctx, f.Team, nil, nil)
		if err != nil {
			return nil, "Unable to fetch Postgres branches."
		}
		return branchSuggestions(instances, args.Get("postgres"), string(f.Environment))
	}
}

func branchSuggestions(instances []postgres.Instance, name, environment string) ([]string, string) {
	if environment == "" {
		envs := postgresEnvironments(instances, name)
		if len(envs) != 1 {
			return nil, "Specify --environment to complete branches for this Postgres."
		}
		environment = envs[0]
	}
	branches := postgresBranches(instances, name, environment)
	if len(branches) == 1 {
		return branches, ""
	}
	return branches, "Select a Postgres branch."
}

func resolvePostgresEnvironment(ctx context.Context, team, name, provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}
	instances, err := postgres.GetTeamPostgresBranches(ctx, team, nil, nil)
	if err != nil {
		return "", fmt.Errorf("fetching environments for Postgres %q: %w", name, err)
	}
	envs := postgresEnvironments(instances, name)
	if len(envs) == 0 {
		return "", fmt.Errorf("Postgres %q not found in team %q", name, team)
	}
	selected, err := input.Select(fmt.Sprintf("Select environment for %s", name), envs, input.SelectWithAutoSelectSingleOption())
	if errors.Is(err, input.ErrNotInteractive) {
		return "", fmt.Errorf("Postgres %q: specify environment with -e, --environment (available: %s)", name, strings.Join(envs, ", "))
	}
	if err != nil {
		return "", fmt.Errorf("selecting environment: %w", err)
	}
	return selected, nil
}
