package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func describeOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Describe{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:         "describe",
		Title:        "Describe an OpenSearch instance.",
		Description:  "This command describes an OpenSearch instance, listing its current configuration and access list.",
		Flags:        flags,
		Args:         defaultArgs,
		ValidateFunc: defaultValidateFunc,
		Examples: []naistrix.Example{
			{
				Description: "Describe an existing OpenSearch instance named some-opensearch for my-team in the dev environment.",
				Command:     "my-team dev some-opensearch",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			pterm.DefaultSection.Println("OpenSearch instance details")
			err = pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(opensearch.FormatDetails(metadata, existing)).
				Render()
			if err != nil {
				return fmt.Errorf("rendering table: %w", err)
			}

			pterm.DefaultSection.Println("OpenSearch access list")
			return pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(opensearch.FormatAccessList(metadata, existing)).
				Render()
		},
	}
}
