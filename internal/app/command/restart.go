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
		Name:  "restart",
		Title: "Restart an application.",
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := app.RestartApp(ctx, flags.Team, args.Get("name"), flags.Environment)
			if err != nil {
				return err
			}

			out.Println(ret)
			return nil
		},
	}
}
