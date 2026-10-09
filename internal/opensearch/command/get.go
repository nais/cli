package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

func get(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Get{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get an OpenSearch instance.",
		Description: "This command describes an OpenSearch instance, listing its current configuration and access list.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name", Prompt: "Name of the OpenSearch instance to inspect"},
		},
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireEnvironment(flags),
			validateArgs,
		),
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

			return renderOpenSearchDetails(out, metadata, existing)
		},
	}
}

func renderOpenSearchDetails(out *naistrix.OutputWriter, metadata opensearch.Metadata, existing *gql.GetOpenSearchTeamEnvironmentOpenSearch) error {
	settings := opensearch.FormatDetails(metadata, existing)[1:]
	width := len("Status:")
	for _, setting := range settings {
		if len(setting[0])+1 > width {
			width = len(setting[0]) + 1
		}
	}

	out.Printf("%s · %s\n", pterm.NewStyle(pterm.Bold).Sprint(metadata.Name), pterm.FgGray.Sprint(metadata.EnvironmentName))
	out.Println(fmt.Sprintf("  %-*s %s", width, "Status:", state(existing.State)))
	out.Println()
	for _, setting := range settings {
		switch setting[0] {
		case "Name", "Environment", "State":
			continue
		}
		out.Printf("  %-*s %s\n", width, setting[0]+":", setting[1])
	}

	out.Println("\nAccess")
	access := opensearch.FormatAccessList(metadata, existing)
	if len(access) == 1 {
		out.Println("No workloads have access.")
		return nil
	}
	if err := out.Table(output.TableWithTopMargin()).Render(access); err != nil {
		return fmt.Errorf("rendering access table: %w", err)
	}
	return nil
}
