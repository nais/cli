package command

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func postgresNames(instances []postgres.PostgresInstance, environment string) []string {
	seen := make(map[string]bool)
	for _, instance := range instances {
		if environment != "" && instance.Environment != environment {
			continue
		}
		seen[instance.Name.Name] = true
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func postgresEnvironments(instances []postgres.PostgresInstance, name string) []string {
	seen := make(map[string]bool)
	for _, instance := range instances {
		if instance.Name.Name == name {
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
		instances, err := postgres.GetTeamPostgreses(ctx, f.Team, nil, nil)
		if err != nil {
			return nil, fmt.Sprintf("Unable to fetch Postgres names: %v", err)
		}
		names := postgresNames(instances, string(f.Environment))
		if len(names) == 1 {
			return names, ""
		}
		return names, "Select a Postgres instance."
	}
}

func branchCompletionEnvironment(instances []postgres.PostgresInstance, name, provided string) (string, string) {
	if provided != "" {
		return provided, ""
	}
	envs := postgresEnvironments(instances, name)
	if len(envs) != 1 {
		return "", "Specify --environment to complete branches for this Postgres."
	}
	return envs[0], ""
}

func resolveBranchCompletionEnvironment(ctx context.Context, team, name, provided string) (string, string) {
	if provided != "" {
		return provided, ""
	}
	instances, err := postgres.GetTeamPostgreses(ctx, team, nil, nil)
	if err != nil {
		return "", fmt.Sprintf("Unable to fetch Postgres environments: %v", err)
	}
	return branchCompletionEnvironment(instances, name, provided)
}

func branchSuggestions(branches []postgres.Branch) ([]string, string) {
	names := make([]string, 0, len(branches))
	for _, branch := range branches {
		names = append(names, branch.Name)
	}
	sort.Strings(names)
	if len(names) == 1 {
		return names, ""
	}
	return names, "Select a Postgres branch."
}

func autoCompletePostgresBranches(f *flag.Postgres) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() == 0 {
			return autoCompletePostgresNames(f)(ctx, args, "")
		}
		if args.Len() != 1 || f.Team == "" {
			return nil, ""
		}
		name := args.Get("postgres")
		env, help := resolveBranchCompletionEnvironment(ctx, f.Team, name, string(f.Environment))
		if help != "" {
			return nil, help
		}
		status, err := postgres.GetBranchStatus(ctx, f.Team, env, name)
		if err != nil {
			return nil, fmt.Sprintf("Unable to fetch Postgres branches: %v", err)
		}
		return branchSuggestions(status.Branches)
	}
}

func resolvePostgresCreateEnvironment(ctx context.Context, provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}
	const hint = "specify an environment using `nais defaults set environment <environment>` or by using the -e, --environment flag"
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching environments: %w", err)
	}
	if len(envs) == 0 {
		return "", fmt.Errorf("missing required environment, %s", hint)
	}
	sort.Strings(envs)
	selected, err := input.Select("Select environment to create Postgres in", envs)
	if errors.Is(err, input.ErrNotInteractive) {
		return "", fmt.Errorf("missing required environment, %s", hint)
	}
	if err != nil {
		return "", err
	}
	return selected, nil
}

func resolvePostgresEnvironment(ctx context.Context, team, name, provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}
	instances, err := postgres.GetTeamPostgreses(ctx, team, nil, nil)
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
