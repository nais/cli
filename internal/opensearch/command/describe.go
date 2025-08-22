package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
)

func describeOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Describe{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "describe",
		Title:       "Describe a OpenSearch instance.",
		Description: "This command describes a OpenSearch instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
