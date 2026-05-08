package command

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/nais/cli/internal/config"
)

func resolveConfigEnvironment(ctx context.Context, team, name, provided string) (string, error) {
	envs, err := config.ConfigEnvironments(ctx, team, name)
	if err != nil {
		return "", fmt.Errorf("fetching environments for config %q: %w", name, err)
	}

	return selectConfigEnvironment(team, name, provided, envs)
}

func selectConfigEnvironment(team, name, provided string, envs []string) (string, error) {
	if provided != "" {
		if slices.Contains(envs, provided) {
			return provided, nil
		}

		if len(envs) == 0 {
			return "", fmt.Errorf("config %q not found in team %q", name, team)
		}

		sort.Strings(envs)
		return "", fmt.Errorf("config %q does not exist in environment %q; available environments: %s", name, provided, strings.Join(envs, ", "))
	}

	switch len(envs) {
	case 0:
		return "", fmt.Errorf("config %q not found in team %q", name, team)
	case 1:
		return envs[0], nil
	default:
		sort.Strings(envs)
		return "", fmt.Errorf("config %q exists in multiple environments (%s); specify -e, --environment", name, strings.Join(envs, ", "))
	}
}
