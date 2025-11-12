package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func issues(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Issues{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "issues",
		Title: "Show issues for an application.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := app.GetApplicationIssues(ctx, flags.Team, args.Get("name"), flags.Environment)
			if err != nil {
				return err
			}
			if len(ret) == 0 {
				out.Println("No issues found for application.")
				return nil
			}

			return out.Table().Render(ret)
		},
	}
}
