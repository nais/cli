package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
)

func listOpenSearches(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.List{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List existing OpenSearch instances.",
		Description: "This command lists all OpenSearch instances across your teams.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// TODO: filter by team and environment
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
