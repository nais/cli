package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	issueutil "github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func issues(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Issues{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "issues",
		Title:       "Show issues for an application.",
		Description: "Lists all active issues detected for a specific application, such as missing resources, configuration errors, or runtime problems.",
		Args: []naistrix.Argument{
			{Name: "name", Prompt: "Name of the application to show issues for", ChoicesFunc: appNameChoices(parentFlags)},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			issueType, err := issueutil.ParseIssueType(flags.IssueType)
			if err != nil {
				return err
			}
			var filter *gql.ResourceIssueFilter
			if issueType != nil {
				filter = &gql.ResourceIssueFilter{IssueType: issueType}
			}
			ret, err := app.GetApplicationIssues(ctx, flags.Team, args.Get("name"), string(flags.Environment), filter)
			if err != nil {
				return err
			}
			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}
			if len(ret) == 0 {
				out.Println("No issues found for application.")
				return nil
			}

			return out.Table().Render(ret)
		},
	}
}
