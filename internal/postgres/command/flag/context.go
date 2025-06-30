package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/v2/internal/k8s"
)

type Context string

func (c *Context) AutoComplete(ctx context.Context, args []string, toComplete string, flags any) ([]string, string) {
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
