package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func createOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Create{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create an OpenSearch instance.",
		Description: "This command creates an OpenSearch instance.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args []string) error {
			if err := flags.Validate(); err != nil {
				return err
			}
			return defaultValidateFunc(ctx, args)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create an OpenSearch instance named some-opensearch for my-team in the dev environment, with default settings.",
				Command:     "my-team dev some-opensearch",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch for my-team in the dev environment, with the specified |SIZE|.",
				Command:     "my-team dev some-opensearch --size RAM_4GB",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch for my-team in the dev environment, with the specified |TIER|.",
				Command:     "my-team dev some-opensearch --tier SINGLE_NODE",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch for my-team in the dev environment, with the specified major |VERSION|.",
				Command:     "my-team dev some-opensearch --version V2",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch for my-team in the dev environment, with all possible options specified.",
				Command:     "my-team dev some-opensearch --size RAM_4GB --tier SINGLE_NODE --version V2",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			// defaults
			data := &opensearch.OpenSearch{
				Size:    gql.OpenSearchSizeRam4gb,
				Tier:    gql.OpenSearchTierSingleNode,
				Version: gql.OpenSearchMajorVersionV2,
			}

			if flags.Size != "" {
				data.Size = gql.OpenSearchSize(flags.Size)
			}
			if flags.Tier != "" {
				data.Tier = gql.OpenSearchTier(flags.Tier)
			}
			if flags.Version != "" {
				data.Version = gql.OpenSearchMajorVersion(flags.Version)
			}

			info := pterm.TableData{
				{"Field", "Value"},
				{"Team", metadata.TeamSlug},
				{"Environment", metadata.EnvironmentName},
				{"Name", metadata.Name},
				{"Size", string(data.Size)},
				{"Tier", string(data.Tier)},
				{"Major version", string(data.Version)},
			}

			pterm.Info.Println("You are about to create an OpenSearch instance with the following configuration:")
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(info).Render(); err != nil {
				return err
			}

			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			_, err := opensearch.Create(ctx, metadata, data)
			if err != nil {
				return err
			}

			pterm.Success.Printfln("Created OpenSearch instance %q for %q in %q", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
