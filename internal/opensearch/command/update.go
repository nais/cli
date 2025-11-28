package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func updateOpenSearch(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Update{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "update",
		Title:       "Update an Opensearch instance.",
		Description: "This command updates an existing Opensearch instance.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := flags.Validate(); err != nil {
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
				Description: "Set the |MEMORY| for an Opensearch instance named some-opensearch.",
				Command:     "some-opensearch --memory GB_8",
			},
			{
				Description: "Set the |TIER| for an OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch --tier HIGH_AVAILABILITY",
			},
			{
				Description: "Set the |STORAGE-GB| for an OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch --storage-gb 1000",
			},
			{
				Description: "Upgrade the major |VERSION| for an OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch --version V2",
			},
			{
				Description: "Set all available options for an OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch --memory GB_8 --tier HIGH_AVAILABILITY --version V2 --storage-gb 1000",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			data := &opensearch.OpenSearch{
				Tier:      existing.Tier,
				Memory:    existing.Memory,
				Version:   existing.Version.DesiredMajor,
				StorageGB: existing.StorageGB,
			}

			outData := pterm.TableData{
				{"Field", "Old Value", "New Value"},
				{"Team", metadata.TeamSlug, "(unchanged)"},
				{"Environment", metadata.EnvironmentName, "(unchanged)"},
				{"Name", metadata.Name, "(unchanged)"},
			}

			newTier := "(unchanged)"
			if flags.Tier != "" && string(flags.Tier) != string(existing.Tier) {
				data.Tier = gql.OpenSearchTier(flags.Tier)
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing tier from %q to %q\n", existing.Tier, data.Tier)
				}
				newTier = string(data.Tier)
			}
			outData = append(outData, []string{"Tier", string(existing.Tier), newTier})

			newMemory := "(unchanged)"
			if flags.Memory != "" && string(flags.Memory) != string(existing.Memory) {
				data.Memory = gql.OpenSearchMemory(flags.Memory)
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing memory from %q to %q\n", existing.Memory, data.Memory)
				}
				newMemory = string(data.Memory)
			}
			outData = append(outData, []string{"Memory", string(existing.Memory), newMemory})

			newMajorVersion := "(unchanged)"
			if flags.MajorVersion != "" && string(flags.MajorVersion) != string(existing.Version.DesiredMajor) {
				data.Version = gql.OpenSearchMajorVersion(flags.MajorVersion)
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing major version from %q to %q\n", existing.Version.DesiredMajor, data.Version)
				}
				newMajorVersion = string(data.Version)
			}
			outData = append(outData, []string{"Major version", string(existing.Version.DesiredMajor), newMajorVersion})

			newStorageGB := "(unchanged)"
			if flags.StorageGB > 0 && flags.StorageGB != existing.StorageGB {
				storage, err := normalizeStorage(data.Tier, data.Memory, flags.StorageGB)
				if err != nil {
					return err
				}
				data.StorageGB = storage

				if flags.IsVerbose() {
					pterm.Info.Printf("Changing storage capacity from %d GB to %d GB\n", existing.StorageGB, data.StorageGB)
				}
				newStorageGB = strconv.Itoa(data.StorageGB)
			}
			outData = append(outData, []string{"Storage capacity", strconv.Itoa(existing.StorageGB), newStorageGB})

			pterm.Info.Println("You are about to update an OpenSearch instance with the following configuration:")
			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(outData).Render(); err != nil {
				return err
			}

			pterm.Warning.Println("Changing settings may cause a restart of the OpenSearch instance.")
			result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
			if !result {
				return fmt.Errorf("cancelled by user")
			}

			_, err = opensearch.Update(ctx, metadata, data)
			if err != nil {
				return err
			}

			pterm.Success.Printf("Updated OpenSearch instance %q for %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
