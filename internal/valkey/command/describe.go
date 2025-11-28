package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func describeValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Describe{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "describe",
		Title:       "Describe a Valkey instance.",
		Description: "This command describes a Valkey instance, listing its current configuration and access list.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}
			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				instances, err := valkey.GetAll(ctx, flags.Team)
				if err != nil {
					return nil, "Unable to fetch Valkey instances."
				}
				var names []string
				for _, instance := range instances {
					names = append(names, instance.Name)
				}
				return names, "Select a Valkey instance."
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Describe an existing Valkey instance named some-valkey.",
				Command:     "some-valkey",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, flags.Team, string(flags.Environment))

			existing, err := valkey.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching existing Valkey instance: %w", err)
			}

			pterm.DefaultSection.Println("Valkey instance details")
			err = pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(valkey.FormatDetails(metadata, existing)).
				Render()
			if err != nil {
				return fmt.Errorf("rendering table: %w", err)
			}

			pterm.DefaultSection.Println("Valkey access list")
			return pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(valkey.FormatAccessList(metadata, existing)).
				Render()
		},
	}
}
