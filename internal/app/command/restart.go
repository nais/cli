package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func restart(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Restart{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "restart",
		Title:       "Restart an application.",
		Description: "Triggers a rolling restart of the application.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			environment, err := resolveAppEnvironment(ctx, out, flags.Team, name, string(flags.Environment), false)
			if err != nil {
				return err
			}

			ret, err := app.RestartApp(ctx, flags.Team, name, environment)
			if err != nil {
				return err
			}

			out.Println(ret)
			return nil
		},
		AutoCompleteFunc: autoCompleteAppNames(flags.App),
	}
}
