package command

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/cli/internal/secret"
)

func resolveSecretEnvironment(ctx context.Context, team, name, provided string) (string, error) {
	envs, err := secret.SecretEnvironments(ctx, team, name)
	if err != nil {
		return "", fmt.Errorf("fetching environments for secret %q: %w", name, err)
	}

	if provided != "" {
		if slices.Contains(envs, provided) {
			return provided, nil
		}

		if len(envs) == 0 {
			return "", fmt.Errorf("secret %q not found in team %q", name, team)
		}

		sort.Strings(envs)
		return "", fmt.Errorf("secret %q does not exist in environment %q; available environments: %s", name, provided, strings.Join(envs, ", "))
	}

	switch len(envs) {
	case 0:
		return "", fmt.Errorf("secret %q not found in team %q", name, team)
	case 1:
		return envs[0], nil
	default:
		sort.Strings(envs)
		return "", fmt.Errorf("secret %q exists in multiple environments (%s); specify --environment/-e", name, strings.Join(envs, ", "))
	}
}

func validateSingleEnvironmentFlagUsage() error {
	if countEnvironmentFlagsInCLIArgs() > 1 {
		return fmt.Errorf("exactly one environment must be specified")
	}
	return nil
}

func countEnvironmentFlagsInCLIArgs() int {
	return cliflags.CountFlagOccurrences(os.Args, "-e", "--environment")
}

func environmentValuesFromCLIArgs() []string {
	return cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
}
