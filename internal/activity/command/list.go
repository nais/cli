package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/activity/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
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
		Name:  "list",
		Title: "List activity for the team.",
		Flags: f,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			activityTypes, err := parseActivityTypes(f.ActivityType)
			if err != nil {
				return err
			}

			resourceTypes, err := parseResourceTypes(f.ResourceType)
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

func parseActivityTypes(in []string) ([]gql.ActivityLogActivityType, error) {
	ret := make([]gql.ActivityLogActivityType, 0, len(in))
	allowed := make(map[string]gql.ActivityLogActivityType, len(gql.AllActivityLogActivityType))
	for _, v := range gql.AllActivityLogActivityType {
		allowed[string(v)] = v
	}

	for _, t := range in {
		normalized := strings.ToUpper(strings.TrimSpace(t))
		v, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("invalid activity type %q", t)
		}
		ret = append(ret, v)
	}

	return ret, nil
}

func parseResourceTypes(in []string) ([]gql.ActivityLogEntryResourceType, error) {
	ret := make([]gql.ActivityLogEntryResourceType, 0, len(in))
	allowed := make(map[string]gql.ActivityLogEntryResourceType, len(gql.AllActivityLogEntryResourceType))
	for _, v := range gql.AllActivityLogEntryResourceType {
		allowed[string(v)] = v
	}

	for _, t := range in {
		normalized := strings.ToUpper(strings.TrimSpace(t))
		v, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("invalid resource type %q", t)
		}
		ret = append(ret, v)
	}

	return ret, nil
}
