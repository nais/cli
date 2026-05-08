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
		Name:        "apply",
		Title:       "Apply resources.",
		Description: "Apply a Nais resource manifest (YAML) to a specific team and environment. The manifest file, team, and environment are required.",
		Args: []naistrix.Argument{
			{Name: "file"},
		},
		AutoCompleteExtensions: []string{"yaml", "yml"},
		Flags:                  flags,
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireTeamAndEnvironment(flags),
			func(ctx context.Context, args *naistrix.Arguments) error {
				if args.Get("file") == "" {
					return fmt.Errorf("file cannot be empty")
				}
				return nil
			},
		),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return apply.Run(ctx, string(flags.Environment), args.Get("file"), flags, out)
		},
	}
}
