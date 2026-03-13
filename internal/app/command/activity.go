package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
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
		Name:  "activity",
		Title: "Show activity for an application.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			activityTypes, err := parseActivityTypes(flags.ActivityType)
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
					return nil, "Please provide team to auto-complete application names. 'nais config set team <team>', or '--team <team>' flag."
				}
				apps, err := app.GetApplicationNames(ctx, flags.Team, flags.Environment)
				if err != nil {
					return nil, "Unable to fetch application names."
				}
				return apps, "Select an application."
			}
			return nil, ""
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
