package command

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/apply"
	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/naistrix"
)

func Apply(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.Apply{Alpha: parentFlags}
	return &naistrix.Command{
		Name:  "apply",
		Title: "Apply resources.",
		Args: []naistrix.Argument{
			{Name: "environment"},
			{Name: "file"},
		},
		AutoCompleteExtensions: []string{"toml"},
		Flags:                  flags,
		ValidateFunc: func(_ context.Context, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("environment cannot be empty")
			}
			if args[1] == "" {
				return fmt.Errorf("file cannot be empty")
			}
			if flags.Team == "" {
				return fmt.Errorf("team cannot be empty")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			environment := args[0]
			filePath := args[1]

			return apply.Run(ctx, environment, filePath, flags, out)
		},
	}
}
