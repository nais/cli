package command

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/apply"
	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/validation"
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
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if args.Get("environment") == "" {
				return fmt.Errorf("environment cannot be empty")
			}
			if args.Get("file") == "" {
				return fmt.Errorf("file cannot be empty")
			}

			return validation.CheckTeam(flags.Team)
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			environment := args.Get("environment")
			filePath := args.Get("file")

			return apply.Run(ctx, environment, filePath, flags, out)
		},
	}
}
