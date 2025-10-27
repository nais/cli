package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func deleteOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Delete{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:         "delete",
		Title:        "Delete an OpenSearch instance.",
		Description:  "This command deletes an existing OpenSearch instance.",
		Flags:        flags,
		Args:         defaultArgs,
		ValidateFunc: defaultValidateFunc,
		Examples: []naistrix.Example{
			{
				Description: "Delete an existing OpenSearch instance named some-opensearch for my-team in the dev environment.",
				Command:     "my-team dev some-opensearch",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args)

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			if len(existing.Access.Edges) > 0 {
				pterm.Error.Println("This OpenSearch instance cannot be deleted as it is currently in use by the following workloads:")
				err = pterm.DefaultTable.
					WithHasHeader().
					WithHeaderRowSeparator("-").
					WithData(opensearch.FormatAccessList(metadata, existing)).
					Render()
				if err != nil {
					return err
				}

				pterm.Info.Println("Remove all references to this OpenSearch instance from the workloads and try again.")
				return errors.New("")
			}

			pterm.Warning.Println("You are about to delete an OpensSarch instance with the following configuration:")
			err = pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(opensearch.FormatDetails(metadata, existing)).
				Render()
			if err != nil {
				return err
			}
			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			deleted, err := opensearch.Delete(ctx, metadata)
			if err != nil {
				return err
			}

			if deleted {
				pterm.Success.Printf("Deleted OpenSearch instance %q from %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			}
			return nil
		},
	}
}
