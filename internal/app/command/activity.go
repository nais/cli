package command

import (
	"context"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func activity(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Activity{
		App:   parentFlags,
		Limit: 20,
	}
	flags.Output = "table"

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

			ret, found, err := app.GetApplicationActivity(ctx, flags.Team, args.Get("name"), string(flags.Environment), activityTypes, flags.Limit)
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
		AutoCompleteFunc: autoCompleteAppNames(parentFlags),
	}
}
