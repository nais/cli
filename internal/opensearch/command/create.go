package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
)

func createOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Create{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create a OpenSearch instance.",
		Description: "This command creates a OpenSearch instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
