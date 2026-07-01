package command

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/nais/cli/internal/app"
	"github.com/nais/naistrix/input"
)

// resolveAppEnvironment determines which environment to use for a command operating on
// a single application.
//
//   - If provided is non-empty, it is validated against the environments where the
//     application exists.
//   - If provided is empty, an interactive selector is shown when running in a
//     terminal. In a non-interactive context (CI, pipes), a clear error is returned
//     listing the available environments.
//   - If the application is not found in any environment, an error is returned.
func resolveAppEnvironment(ctx context.Context, team, name, provided string) (string, error) {
	envs, err := app.ApplicationEnvironments(ctx, team, name)
	if err != nil {
		return "", fmt.Errorf("fetching environments for application %q: %w", name, err)
	}

	if provided != "" {
		if slices.Contains(envs, provided) {
			return provided, nil
		}
		if len(envs) == 0 {
			return "", fmt.Errorf("application %q not found in team %q", name, team)
		}
		sort.Strings(envs)
		return "", fmt.Errorf("application %q does not exist in environment %q; available environments: %s", name, provided, strings.Join(envs, ", "))
	}

	if len(envs) == 0 {
		return "", fmt.Errorf("application %q not found in team %q", name, team)
	}

	sort.Strings(envs)
	selected, err := input.Select(fmt.Sprintf("Select environment for %s", name), envs)
	if errors.Is(err, input.ErrNotInteractive) {
		return "", fmt.Errorf("application %q: specify environment with -e, --environment (available: %s)", name, strings.Join(envs, ", "))
	}
	if err != nil {
		return "", fmt.Errorf("selecting environment: %w", err)
	}
	return selected, nil
}
