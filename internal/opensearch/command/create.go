package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
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
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := flags.Validate(); err != nil {
				return err
			}

			if err := validateArgs(args); err != nil {
				return err
			}

			return validation.CheckTeam(flags.Team)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create an OpenSearch instance named some-opensearch in the dev environment, with default settings.",
				Command:     "dev some-opensearch",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch in the dev environment, with the specified |MEMORY|.",
				Command:     "dev some-opensearch --memory GB_4",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch in the dev environment, with the specified |TIER|.",
				Command:     "dev some-opensearch --tier SINGLE_NODE",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch in the dev environment, with the specified major |VERSION|.",
				Command:     "dev some-opensearch --version V2",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch in the dev environment, with the specified |STORAGE-GB|.",
				Command:     "dev some-opensearch --storage-gb 100",
			},
			{
				Description: "Create an OpenSearch instance named some-opensearch in the dev environment, with all possible options specified.",
				Command:     "dev some-opensearch --memory GB_4 --tier SINGLE_NODE --version V2 --storage-gb 100",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team)

			// defaults
			data := &opensearch.OpenSearch{
				Tier:      gql.OpenSearchTierSingleNode,
				Memory:    gql.OpenSearchMemoryGb4,
				StorageGB: 0,
				Version:   gql.OpenSearchMajorVersionV2,
			}

			if flags.Tier != "" {
				data.Tier = gql.OpenSearchTier(flags.Tier)
			}
			if flags.Memory != "" {
				data.Memory = gql.OpenSearchMemory(flags.Memory)
			}
			if flags.StorageGB != 0 {
				data.StorageGB = flags.StorageGB
			}
			if flags.Version != "" {
				data.Version = gql.OpenSearchMajorVersion(flags.Version)
			}

			storage, err := normalizeStorage(data.Tier, data.Memory, data.StorageGB)
			if err != nil {
				return err
			}
			data.StorageGB = storage

			info := pterm.TableData{
				{"Field", "Value"},
				{"Team", metadata.TeamSlug},
				{"Environment", metadata.EnvironmentName},
				{"Name", metadata.Name},
				{"Tier", string(data.Tier)},
				{"Memory", string(data.Memory)},
				{"Storage", fmt.Sprintf("%d GB", data.StorageGB)},
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

			_, err = opensearch.Create(ctx, metadata, data)
			if err != nil {
				return err
			}

			pterm.Success.Printfln("Created OpenSearch instance %q for %q in %q", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
