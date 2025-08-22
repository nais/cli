package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func createValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Create{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create a Valkey instance.",
		Description: "This command creates a Valkey instance.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
			{Name: "team"},
			{Name: "environment"},
		},
		ValidateFunc: func(_ context.Context, args []string) error {
			if len(args) != 3 {
				return fmt.Errorf("expected 3 arguments, got %d", len(args))
			}
			if args[0] == "" {
				return fmt.Errorf("name cannot be empty")
			}
			if args[1] == "" {
				return fmt.Errorf("team cannot be empty")
			}
			if args[2] == "" {
				return fmt.Errorf("environment cannot be empty")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			metadata := valkey.Metadata{
				Name:            args[0],
				TeamSlug:        args[1],
				EnvironmentName: args[2],
			}

			// defaults
			data := &valkey.Valkey{
				Size: "RAM_4GB",
				Tier: "SINGLE_NODE",
			}

			if flags.Size != "" {
				data.Size = gql.ValkeySize(flags.Size)
			}
			if flags.Tier != "" {
				data.Tier = gql.ValkeyTier(flags.Tier)
			}

			_, err := valkey.Create(ctx, metadata, data)
			return err
		},
		// TODO: completion, examples, etc.
		//  how do we generate valid options for size and tier in usage text?
		//  how do we display defaults for size and tier?
		//  should team and environment be flags? default to some stored state for the current authenticated user?
	}
}
