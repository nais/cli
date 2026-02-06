package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Environment string

func (e *Environment) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	return autoCompleteEnvironments(ctx)
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}

	return envs, "Available environments"
}
