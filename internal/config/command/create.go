package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func create(parentFlags *flag.Config) *naistrix.Command {
	f := &flag.Create{Config: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create a new config.",
		Description: "This command creates a new empty config in a team environment.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			return validateArgs(args)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create a config named my-config in environment dev.",
				Command:     "my-config --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			_, err := config.Create(ctx, metadata)
			if err != nil {
				return fmt.Errorf("creating config: %w", err)
			}

			pterm.Success.Printfln("Created config %q in %q for team %q", metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
			return nil
		},
	}
}
