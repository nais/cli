package command

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nais/cli/internal/secret"
)

func resolveSecretEnvironment(ctx context.Context, team, name, provided string) (string, error) {
	envs, err := secret.SecretEnvironments(ctx, team, name)
	if err != nil {
		return "", fmt.Errorf("fetching environments for secret %q: %w", name, err)
	}

	if provided != "" {
		for _, env := range envs {
			if env == provided {
				return provided, nil
			}
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
	count := 0
	args := os.Args

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-e" || arg == "--environment":
			if i+1 >= len(args) {
				continue
			}
			next := args[i+1]
			if strings.HasPrefix(next, "-") || next == "" {
				continue
			}
			count++
			i++
		case strings.HasPrefix(arg, "--environment="):
			if strings.TrimPrefix(arg, "--environment=") != "" {
				count++
			}
		case strings.HasPrefix(arg, "-e="):
			if strings.TrimPrefix(arg, "-e=") != "" {
				count++
			}
		}
	}

	return count
}

func environmentValuesFromCLIArgs() []string {
	seen := map[string]struct{}{}
	environments := make([]string, 0)
	args := os.Args

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-e" || arg == "--environment":
			if i+1 >= len(args) {
				continue
			}
			next := args[i+1]
			if strings.HasPrefix(next, "-") || next == "" {
				continue
			}
			if _, ok := seen[next]; !ok {
				seen[next] = struct{}{}
				environments = append(environments, next)
			}
			i++
		case strings.HasPrefix(arg, "--environment="):
			env := strings.TrimPrefix(arg, "--environment=")
			if env != "" {
				if _, ok := seen[env]; !ok {
					seen[env] = struct{}{}
					environments = append(environments, env)
				}
			}
		case strings.HasPrefix(arg, "-e="):
			env := strings.TrimPrefix(arg, "-e=")
			if env != "" {
				if _, ok := seen[env]; !ok {
					seen[env] = struct{}{}
					environments = append(environments, env)
				}
			}
		}
	}

	return environments
}
