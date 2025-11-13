package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
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

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				apps, err := app.GetApplicationNames(ctx, flags.Team)
				if err != nil {
					return nil, "Unable to fetch application names."
				}
				return apps, "Select an application."
			}
			return nil, ""
		},
	}
}
