package command

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/apply"
	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Apply(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Apply{GlobalFlags: parentFlags, Timeout: 10 * time.Minute}
	return &naistrix.Command{
		Name:        "apply",
		Title:       "Apply resources.",
		Description: "Apply Nais resource manifests (YAML) to a specific team and environment. Accepts a single file or a directory containing multiple manifests. When a directory is given, mixin files (<base>.<env>.yaml) are auto-loaded and --set/--append/--mixin flags are disabled.",
		Examples: []naistrix.Example{
			{
				Description: "Replace the environment variable list.",
				Command:     "app.yaml --environment dev-gcp --team nais --set 'spec.env=[{name: LOG_LEVEL, value: debug}]'",
			},
			{
				Description: "Update a field in an existing list element.",
				Command:     "app.yaml --environment dev-gcp --team nais --set 'spec.env[0].value=debug'",
			},
			{
				Description: "Append an environment variable without replacing existing entries.",
				Command:     "app.yaml --environment dev-gcp --team nais --append 'spec.env={name: LOG_LEVEL, value: debug}'",
			},
		},
		Args: []naistrix.Argument{
			{Name: "path"},
		},
		AutoCompleteExtensions: []string{"yaml", "yml"},
		Flags:                  flags,
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireTeam(flags),
			func(ctx context.Context, args *naistrix.Arguments) error {
				if args.Get("path") == "" {
					return fmt.Errorf("path cannot be empty")
				}
				return nil
			},
		),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return apply.Run(ctx, args.Get("path"), flags, out)
		},
	}
}
