package flags

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Environment string

var _ naistrix.FlagAutoCompleter = (*Environment)(nil)

func (e *Environment) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, _ any) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}
