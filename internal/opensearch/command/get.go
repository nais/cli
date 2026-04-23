package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func get(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Get{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get an OpenSearch instance.",
		Description: "This command describes an OpenSearch instance, listing its current configuration and access list.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			err := validation.CheckEnvironment(string(flags.Environment))
			if err != nil {
				return err
			}
			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteOpenSearchNames(ctx, flags.Team, string(flags.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Describe an existing OpenSearch instance named some-opensearch in environment dev.",
				Command:     "some-opensearch --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			out.Println("OpenSearch instance details")
			if err = out.Table(output.TableWithMargins()).Render(opensearch.FormatDetails(metadata, existing)); err != nil {
				return fmt.Errorf("rendering table: %w", err)
			}

			out.Println("OpenSearch access list")
			return out.Table(output.TableWithTopMargin()).Render(opensearch.FormatAccessList(metadata, existing))
		},
	}
}
