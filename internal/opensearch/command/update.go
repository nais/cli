package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args []string) error {
			if err := flags.Validate(); err != nil {
				return err
			}
			return defaultValidateFunc(ctx, args)
		},
		Examples: []naistrix.Example{
			{
				Description: "Set the |SIZE| for an Opensearch instance named some-opensearch for my-team in the dev environment.",
				Command:     "my-team dev some-opensearch --size RAM_8GB",
			},
			{
				Description: "Set the |TIER| for an OpenSearch instance named some-opensearch for my-team in the dev environment.",
				Command:     "my-team dev some-opensearch --tier HIGH_AVAILABILITY",
			},
			{
				Description: "Upgrade the major |VERSION| for an OpenSearch instance named some-opensearch for my-team in the dev environment.",
				Command:     "my-team dev some-opensearch --version V2",
			},
			{
				Description: "Set all available options for an OpenSearch instance named some-opensearch for my-team in the dev environment.",
				Command:     "my-team dev some-opensearch --size RAM_8GB --tier HIGH_AVAILABILITY --version V2",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := metadataFromArgs(args)

			existing, err := opensearch.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing OpenSearch instance: %w", err)
			}

			data := &opensearch.OpenSearch{
				Size:    existing.Size,
				Tier:    existing.Tier,
				Version: existing.MajorVersion,
			}

			outData := pterm.TableData{
				{"Field", "Old Value", "New Value"},
				{"Team", metadata.TeamSlug, "(unchanged)"},
				{"Environment", metadata.EnvironmentName, "(unchanged)"},
				{"Name", metadata.Name, "(unchanged)"},
			}

			newSize := "(unchanged)"
			if flags.Size != "" && string(flags.Size) != string(existing.Size) {
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing size from %q to %q\n", existing.Size, flags.Size)
				}
				data.Size = gql.OpenSearchSize(flags.Size)
				newSize = string(flags.Size)
			}
			outData = append(outData, []string{"Size", string(existing.Size), newSize})

			newTier := "(unchanged)"
			if flags.Tier != "" && string(flags.Tier) != string(existing.Tier) {
				if flags.IsVerbose() {
					pterm.Info.Printf("Changing tier from %q to %q\n", existing.Tier, data.Tier)
				}
				data.Tier = gql.OpenSearchTier(flags.Tier)
				newTier = string(flags.Tier)
			}
			outData = append(outData, []string{"Tier", string(existing.Tier), newTier})

			newMajorVersion := "(unchanged)"
			if flags.MajorVersion != "" && string(flags.MajorVersion) != string(existing.MajorVersion) {
				if err := validateMajorVersionFlag(string(existing.MajorVersion), string(flags.MajorVersion)); err != nil {
					return err
				}

				if flags.IsVerbose() {
					pterm.Info.Printf("Changing major version from %q to %q\n", existing.MajorVersion, data.Version)
				}
				data.Version = gql.OpenSearchMajorVersion(flags.MajorVersion)
				newMajorVersion = string(flags.MajorVersion)
			}
			outData = append(outData, []string{"Major version", string(existing.MajorVersion), newMajorVersion})

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

func validateMajorVersionFlag(current, desired string) error {
	c, err := strconv.Atoi(strings.TrimPrefix(current, "V"))
	if err != nil {
		return fmt.Errorf("parsing current major version %q: %w", current, err)
	}

	d, err := strconv.Atoi(strings.TrimPrefix(desired, "V"))
	if err != nil {
		return fmt.Errorf("parsing desired major version %q: %w", desired, err)
	}

	if d < c {
		return fmt.Errorf("downgrading major version from %q to %q is not supported", current, desired)
	}
	return nil
}
