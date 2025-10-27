package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/naistrix"
)

type Namespace string

var _ naistrix.FlagAutoCompleter = (*Namespace)(nil)

func (c *Namespace) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, flags any) ([]string, string) {
	f, ok := flags.(*Postgres)
	if !ok {
		return nil, "Invalid flags type"
	}

	namespaceForContext, err := k8s.GetNamespaceForContext(string(f.Context))
	if err != nil {
		return nil, fmt.Sprintf("Error fetching namespace for context %s: %v", f.Context, err)
	}
	f.Namespace = Namespace(namespaceForContext)

	namespaces, err := k8s.GetNamespacesForContext(ctx, string(f.Context))
	if err != nil {
		return nil, fmt.Sprintf("Error fetching namespaces: %v", err)
	}

	return namespaces, "Available namespaces."
}
