package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
	"k8s.io/apimachinery/pkg/api/resource"
)

func update(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Update{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "update",
		Title:       "Update an Opensearch instance.",
		Description: "This command updates an existing Opensearch instance.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireEnvironment(flags),
			validateArgs,
			func(ctx context.Context, args *naistrix.Arguments) error {
				if err := flags.Validate(); err != nil {
					return err
				}
				return nil
			},
		),
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteOpenSearchNames(ctx, flags.Team, string(flags.Environment), true)
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
				Description: "Upgrade the |VERSION| for an OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch --version V3_3",
			},
			{
				Description: "Set all available options for an OpenSearch instance named some-opensearch.",
				Command:     "some-opensearch --memory GB_8 --tier HIGH_AVAILABILITY --version V3_3 --storage-gb 1000 --shard-indexing-pressure-enabled true --shard-indexing-pressure-enforced false --indices-query-bool-max-clause-count 2048 --http-max-content-length 200Mi",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			data := &gql.UpdateOpenSearchInput{
				Tier:                           existing.Tier,
				Memory:                         existing.Memory,
				Version:                        existing.Version.DesiredMajor,
				StorageGB:                      existing.StorageGB,
				ShardIndexingPressureEnabled:   new(existing.ShardIndexingPressureEnabled),
				ShardIndexingPressureEnforced:  new(existing.ShardIndexingPressureEnforced),
				IndicesQueryBoolMaxClauseCount: existing.IndicesQueryBoolMaxClauseCount,
				HttpMaxContentLength:           existing.HttpMaxContentLength,
			}

			outData := [][]string{
				{"Field", "Old Value", "New Value"},
				{"Team", metadata.TeamSlug, "(unchanged)"},
				{"Environment", metadata.EnvironmentName, "(unchanged)"},
				{"Name", metadata.Name, "(unchanged)"},
			}

			newTier := "(unchanged)"
			if flags.Tier != "" && string(flags.Tier) != string(existing.Tier) {
				data.Tier = gql.OpenSearchTier(flags.Tier)
				if flags.IsVerbose() {
					out.Infof("Changing tier from %q to %q\n", existing.Tier, data.Tier)
				}
				newTier = string(data.Tier)
			}
			outData = append(outData, []string{"Tier", string(existing.Tier), newTier})

			newMemory := "(unchanged)"
			if flags.Memory != "" && string(flags.Memory) != string(existing.Memory) {
				data.Memory = gql.OpenSearchMemory(flags.Memory)
				if flags.IsVerbose() {
					out.Infof("Changing memory from %q to %q\n", existing.Memory, data.Memory)
				}
				newMemory = string(data.Memory)
			}
			outData = append(outData, []string{"Memory", string(existing.Memory), newMemory})

			newMajorVersion := "(unchanged)"
			if flags.MajorVersion != "" && string(flags.MajorVersion) != string(existing.Version.DesiredMajor) {
				data.Version = gql.OpenSearchMajorVersion(flags.MajorVersion)
				if flags.IsVerbose() {
					out.Infof("Changing version from %q to %q\n", existing.Version.DesiredMajor, data.Version)
				}
				newMajorVersion = string(data.Version)
			}
			outData = append(outData, []string{"Version", string(existing.Version.DesiredMajor), newMajorVersion})

			newStorageGB := "(unchanged)"
			if flags.StorageGB > 0 && flags.StorageGB != existing.StorageGB {
				storage, err := normalizeStorage(data.Tier, data.Memory, flags.StorageGB)
				if err != nil {
					return err
				}
				data.StorageGB = storage

				if flags.IsVerbose() {
					out.Infof("Changing storage capacity from %d GB to %d GB\n", existing.StorageGB, data.StorageGB)
				}
				newStorageGB = strconv.Itoa(data.StorageGB)
			}
			outData = append(outData, []string{"Storage capacity", strconv.Itoa(existing.StorageGB), newStorageGB})

			newShardIndexingPressureEnabled := "(unchanged)"
			if flags.ShardIndexingPressureEnabled != "" {
				value, err := flags.ShardIndexingPressureEnabled.Bool()
				if err != nil {
					return err
				}
				if *value != existing.ShardIndexingPressureEnabled {
					data.ShardIndexingPressureEnabled = value
					newShardIndexingPressureEnabled = strconv.FormatBool(*value)
				}
			}
			outData = append(outData, []string{
				"Shard indexing pressure enabled",
				strconv.FormatBool(existing.ShardIndexingPressureEnabled),
				newShardIndexingPressureEnabled,
			})

			newShardIndexingPressureEnforced := "(unchanged)"
			if flags.ShardIndexingPressureEnforced != "" {
				value, err := flags.ShardIndexingPressureEnforced.Bool()
				if err != nil {
					return err
				}
				if *value != existing.ShardIndexingPressureEnforced {
					data.ShardIndexingPressureEnforced = value
					newShardIndexingPressureEnforced = strconv.FormatBool(*value)
				}
			}
			outData = append(outData, []string{
				"Shard indexing pressure enforced",
				strconv.FormatBool(existing.ShardIndexingPressureEnforced),
				newShardIndexingPressureEnforced,
			})

			newIndicesQueryBoolMaxClauseCount := "(unchanged)"
			if flags.IndicesQueryBoolMaxClauseCount != 0 && flags.IndicesQueryBoolMaxClauseCount != optionalIntValue(existing.IndicesQueryBoolMaxClauseCount) {
				data.IndicesQueryBoolMaxClauseCount = new(flags.IndicesQueryBoolMaxClauseCount)
				newIndicesQueryBoolMaxClauseCount = strconv.Itoa(flags.IndicesQueryBoolMaxClauseCount)
			}
			outData = append(outData, []string{
				"Indices query bool max clause count",
				optionalInt(existing.IndicesQueryBoolMaxClauseCount),
				newIndicesQueryBoolMaxClauseCount,
			})

			newHTTPMaxContentLength := "(unchanged)"
			if flags.HttpMaxContentLength != "" {
				quantity, err := resource.ParseQuantity(flags.HttpMaxContentLength)
				if err != nil {
					return fmt.Errorf("parsing HTTP max content length: %w", err)
				}
				httpMaxContentLength := quantity.String()
				if httpMaxContentLength != optionalStringValue(existing.HttpMaxContentLength) {
					data.HttpMaxContentLength = &httpMaxContentLength
					newHTTPMaxContentLength = httpMaxContentLength
				}
			}
			outData = append(outData, []string{
				"HTTP max content length",
				optionalString(existing.HttpMaxContentLength),
				newHTTPMaxContentLength,
			})

			out.Infoln("You are about to update an OpenSearch instance with the following configuration:")
			if err := out.Table(output.TableWithMargins()).Render(outData); err != nil {
				return err
			}

			out.Warnln("Changing settings may cause a restart of the OpenSearch instance.")
			if result, err := input.Confirm("Are you sure you want to continue?"); err != nil {
				return err
			} else if !result {
				return fmt.Errorf("cancelled by user")
			}

			if _, err = opensearch.Update(ctx, metadata, *data); err != nil {
				return err
			}

			out.Successf("Updated OpenSearch instance %q for %q in %q\n", metadata.Name, metadata.TeamSlug, metadata.EnvironmentName)
			return nil
		},
	}
}
