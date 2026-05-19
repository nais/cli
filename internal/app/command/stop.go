package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func stop(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Restart{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "stop",
		Title:       "Stop an application.",
		Description: "Stops an application by setting replicas to 0. Changes are temporary and will be overwritten on next deploy.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		ValidateFunc: func(_ context.Context, _ *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("environment must be specified (-e, --environment)")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			if result, err := input.Confirm(fmt.Sprintf("Are you sure you want to stop %v in %v?", name, flags.Environment)); err != nil {
				return err
			} else if !result {
				return fmt.Errorf("cancelled by user")
			}

			ret, err := app.SetReplicas(ctx, flags.Team, name, string(flags.Environment), 0, 0)
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
