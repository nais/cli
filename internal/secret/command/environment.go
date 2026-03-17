package command

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/nais/cli/internal/cliflags"
)

func selectSecretEnvironment(team, name, provided string, envs []string) (string, error) {
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
		return "", fmt.Errorf("secret %q exists in multiple environments (%s); specify --environment/-e on the command, e.g. nais secrets get -t %s %s -e <%s>", name, strings.Join(envs, ", "), team, name, envs[0])
	}
}

func validateSingleEnvironmentFlagUsage() error {
	if countEnvironmentFlagsInCLIArgs() > 1 {
		return fmt.Errorf("only one --environment/-e flag may be provided")
	}
	return nil
}

func countEnvironmentFlagsInCLIArgs() int {
	return cliflags.CountFlagOccurrences(os.Args, "-e", "--environment")
}

func environmentValuesFromCLIArgs() []string {
	return cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
}
