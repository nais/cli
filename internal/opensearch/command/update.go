package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
)

func updateOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Update{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "update",
		Title:       "Update a OpenSearch instance.",
		Description: "This command updates an existing OpenSearch instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
