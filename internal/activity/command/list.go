package command

import (
	"context"
	"slices"

	"github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/activity/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func list(parentFlags *flag.Activity) *naistrix.Command {
	f := &flag.List{
		Activity: parentFlags,
		Output:   "table",
		Limit:    20,
	}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List activity for the team.",
		Description: "Shows recent events for the team, including deployments and resource changes. Results can be filtered by activity type and resource type.",
		Flags:       f,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			activityTypes, err := activity.ParseActivityTypes(f.ActivityType)
			if err != nil {
				return err
			}

			resourceTypes, err := activity.ParseResourceTypes(f.ResourceType)
			if err != nil {
				return err
			}

			fetchLimit := f.Limit
			if len(resourceTypes) > 0 && fetchLimit < 1000 {
				fetchLimit = 1000
			}

			ret, err := activity.List(ctx, f.Team, activityTypes, fetchLimit)
			if err != nil {
				return err
			}

			if len(resourceTypes) > 0 {
				requested := make([]string, 0, len(resourceTypes))
				for _, v := range resourceTypes {
					requested = append(requested, string(v))
				}

				filtered := make([]activity.Entry, 0, len(ret))
				for _, entry := range ret {
					if slices.Contains(requested, entry.ResourceType) {
						filtered = append(filtered, entry)
					}
				}
				ret = filtered
			}

			if f.Limit > 0 && len(ret) > f.Limit {
				ret = ret[:f.Limit]
			}

			if len(ret) == 0 {
				out.Println("No activity found.")
				return nil
			}

			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
	}
}
