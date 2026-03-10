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
			{Name: "file"},
		},
		AutoCompleteExtensions: []string{"yaml", "yml"},
		Flags:                  flags,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if args.Get("file") == "" {
				return fmt.Errorf("file cannot be empty")
			}

			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}

			if err := validation.CheckTeam(flags.Team); err != nil {
				return err
			}

			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return apply.Run(ctx, string(flags.Environment), args.Get("file"), flags, out)
		},
	}
}
