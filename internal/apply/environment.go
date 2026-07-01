package apply

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

// resolveEnvironment determines which environment to apply to.
//
//   - If provided is non-empty, it is used as-is.
//   - Otherwise, when running in an interactive terminal, an interactive
//     selector lists the available environments.
//   - In a non-interactive context (CI, pipes), a clear error is returned
//     telling the user how to specify an environment.
func resolveEnvironment(ctx context.Context, provided string, out *naistrix.OutputWriter) (string, error) {
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

	selected, err := input.Select("Select environment to apply to", envs)
	if errors.Is(err, input.ErrNotInteractive) {
		return "", fmt.Errorf("missing required environment, %s", hint)
	}
	if err != nil {
		return "", err
	}
	return selected, nil
}
