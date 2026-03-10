package flag

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Env string

func (e *Env) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, _ any) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Apply struct {
	*alpha.Alpha
	Environment Env `name:"environment" short:"e" usage:"The |ENVIRONMENT| to apply resources to."`
}
