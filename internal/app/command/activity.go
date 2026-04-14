package command

import (
	"context"
	"os"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func activity(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Activity{
		App:    parentFlags,
		Output: "table",
		Limit:  20,
	}

	return &naistrix.Command{
		Name:        "activity",
		Title:       "Show activity for an application.",
		Description: "Displays recent events for a specific application, such as deployments and configuration changes. Results can be filtered by activity type.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			activityTypes, err := activityutil.ParseActivityTypes(flags.ActivityType)
			if err != nil {
				return err
			}

			ret, found, err := app.GetApplicationActivity(ctx, flags.Team, args.Get("name"), flags.Environment, activityTypes, flags.Limit)
			if err != nil {
				return err
			}
			if !found {
				out.Println("Application not found.")
				return nil
			}
			if len(ret) == 0 {
				out.Println("No activity found for application.")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if len(flags.Team) == 0 {
					return nil, "Please provide team to auto-complete application names. 'nais defaults set team <team>', or '--team <team>' flag."
				}
				environments := []string(flags.Environment)
				if len(environments) == 0 {
					environments = cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
				}

				apps, err := app.GetApplicationNames(ctx, flags.Team, environments)
				if err != nil {
					return nil, "Unable to fetch application names."
				}
				return apps, "Select an application."
			}
			return nil, ""
		},
	}
}
