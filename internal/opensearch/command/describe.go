package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func describeOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Describe{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "describe",
		Title:       "Describe an OpenSearch instance.",
		Description: "This command describes an OpenSearch instance, listing its current configuration and access list.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			err := validation.CheckEnvironment(string(flags.Environment))
			if err != nil {
				return err
			}
			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				instances, err := opensearch.GetAll(ctx, flags.Team)
				if err != nil {
					return nil, "Unable to fetch OpenSearch instances."
				}
				var names []string
				for _, instance := range instances {
					names = append(names, instance.Name)
				}
				return names, "Select an OpenSearch instance."
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Describe an existing OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

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
