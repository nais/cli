package command

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/nais/cli/internal/app"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

// resolveAppEnvironment determines which environment to use for a command operating on
// a single application.
//
//   - If provided is non-empty, it is validated against the environments where the
//     application exists.
//   - If provided is empty and the application exists in exactly one environment,
//     that environment is auto-selected and a notice is printed via out, unless
//     quiet is true (e.g. when machine-readable output like JSON is requested).
//   - If provided is empty and the application exists in multiple environments,
//     an interactive selector is shown when running in a terminal. In a non-interactive
//     context (CI, pipes), a clear error is returned listing the available environments.
//   - If the application is not found in any environment, an error is returned.
func resolveAppEnvironment(ctx context.Context, out *naistrix.OutputWriter, team, name, provided string, quiet bool) (string, error) {
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

	switch len(envs) {
	case 0:
		return "", fmt.Errorf("application %q not found in team %q", name, team)
	case 1:
		if !quiet {
			out.Infof("%s only found in %s, auto-selecting.\n", name, envs[0])
		}
		return envs[0], nil
	default:
		sort.Strings(envs)
		// Only prompt when both stdin and stdout are terminals: stdin so we can
		// read the user's choice, stdout so the user actually sees the prompt.
		// In pipelines like `nais app status <app> | jq ...`, stdout is a pipe
		// and we must fail clearly instead of blocking on hidden input.
		if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) { // #nosec G115
			return "", fmt.Errorf("application %q exists in multiple environments (%s); specify -e, --environment", name, strings.Join(envs, ", "))
		}
		selected, err := input.Select(fmt.Sprintf("Select environment for %s", name), envs)
		if err != nil {
			return "", fmt.Errorf("selecting environment: %w", err)
		}
		return selected, nil
	}
}
