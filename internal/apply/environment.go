package apply

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
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

	// Only prompt when both stdin and stdout are terminals: stdin so we can read
	// the user's choice, stdout so the user actually sees the prompt.
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) { // #nosec G115
		return "", fmt.Errorf("missing required environment, %s", hint)
	}

	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching environments: %w", err)
	}
	if len(envs) == 0 {
		return "", fmt.Errorf("missing required environment, %s", hint)
	}
	sort.Strings(envs)

	selected, err := input.Select("Select environment to apply to", envs)
	if err != nil {
		return "", err
	}
	return selected, nil
}
