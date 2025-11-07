package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func issues(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.ListApps{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "issues",
		Title: "Show issues for an application.",
		Args: []naistrix.Argument{
			{Name: "name"},
			{Name: "team"},
			{Name: "env"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			teamSlug := args.Get("team")
			name := args.Get("name")
			env := args.Get("env")
			ret, err := app.GetApplicationIssues(ctx, teamSlug, name, env)
			if err != nil {
				return err
			}

			return out.Table().Render(ret)
		},
	}
}
