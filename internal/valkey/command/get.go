package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

func get(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Describe{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get a Valkey instance.",
		Description: "This command describes a Valkey instance, listing its current configuration and access list.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireEnvironment(flags),
			validateArgs,
		),
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteValkeyNames(ctx, flags.Team, string(flags.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Describe an existing Valkey instance named some-valkey in environment dev.",
				Command:     "some-valkey --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			return renderValkeyDetails(out, metadata, existing)
		},
	}
}

func renderValkeyDetails(out *naistrix.OutputWriter, metadata valkey.Metadata, existing *gql.GetValkeyTeamEnvironmentValkey) error {
	settings := valkey.FormatDetails(metadata, existing)[1:]
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
	access := valkey.FormatAccessList(metadata, existing)
	if len(access) == 1 {
		out.Println("No workloads have access.")
		return nil
	}
	if err := out.Table(output.TableWithTopMargin()).Render(access); err != nil {
		return fmt.Errorf("rendering access table: %w", err)
	}
	return nil
}
