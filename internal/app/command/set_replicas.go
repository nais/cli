package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func setReplicas(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.SetReplicas{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "replicas",
		Title:       "Set replica count for an application.",
		Description: "Updates the minimum and maximum replica count for the application. Changes are temporary and will be overwritten on next deploy.",
		Examples: []naistrix.Example{
			{
				Description: "Set replicas to min 2, max 5.",
				Command:     "my-app --min 2 --max 5 -e dev",
			},
		},
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		ValidateFunc: func(_ context.Context, _ *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("environment must be specified (-e, --environment)")
			}
			if flags.Min <= 0 {
				return fmt.Errorf("--min must be a positive integer")
			}
			if flags.Max <= 0 {
				return fmt.Errorf("--max must be a positive integer")
			}
			if flags.Min > flags.Max {
				return fmt.Errorf("--min (%d) cannot be greater than --max (%d)", flags.Min, flags.Max)
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			ret, err := app.SetReplicas(ctx, flags.Team, name, string(flags.Environment), flags.Min, flags.Max)
			if err != nil {
				return err
			}

			out.Println(ret)
			out.Warnf("Changes are temporary and will be overwritten on next deploy.\n")
			return nil
		},
		AutoCompleteFunc: autoCompleteAppNames(parentFlags),
	}
}
