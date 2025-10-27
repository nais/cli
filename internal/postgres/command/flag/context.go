package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/naistrix"
)

type Context string

var _ naistrix.FlagAutoCompleter = (*Context)(nil)

func (c *Context) AutoComplete(_ context.Context, _ *naistrix.Arguments, _ string, flags any) ([]string, string) {
	_, ok := flags.(*Postgres)
	if !ok {
		return nil, "Invalid flags type"
	}

	contexts, err := k8s.GetAllContexts()
	if err != nil {
		return nil, fmt.Sprintf("Error fetching contexts: %v", err)
	}

	return contexts, "Available contexts."
}
