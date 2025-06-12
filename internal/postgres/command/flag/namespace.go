package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
)

type Namespace string

func (c *Namespace) AutoComplete(ctx context.Context, _ []string, _ string, flags any) ([]string, string) {
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
