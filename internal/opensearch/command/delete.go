package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

func delete(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Delete{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete an OpenSearch instance.",
		Description: "This command deletes an existing OpenSearch instance.",
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
				Description: "Delete an existing OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			if len(existing.Access.Edges) > 0 {
				out.Errorln("This OpenSearch instance cannot be deleted as it is currently in use by the following workloads:")
				if err := out.Table(output.TableWithMargins()).Render(opensearch.FormatAccessList(metadata, existing)); err != nil {
					return err
				}

				out.Infoln("Remove all references to this OpenSearch instance from the workloads and try again.")
				return nil
			}

			out.Warnln("You are about to delete an OpenSearch instance with the following configuration:")
			if err := out.Table(output.TableWithMargins()).Render(opensearch.FormatDetails(metadata, existing)); err != nil {
				return err
			}

			if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
				return err
			} else if !result {
				return fmt.Errorf("cancelled by user")
			}

			if deleted, err := opensearch.Delete(ctx, metadata); err != nil {
				return err
			} else if !deleted {
				return fmt.Errorf("OpenSearch instance was not deleted")
			}

			out.Successf("Deleted OpenSearch instance %q from %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
