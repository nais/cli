package command

import (
	"context"
	"os"
	"strings"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
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
					return nil, "Please provide team to auto-complete application names. 'nais config set team <team>', or '--team <team>' flag."
				}
				environments := flags.Environment
				if len(environments) == 0 {
					environments = environmentsFromCLIArgs()
				}
				if len(environments) == 0 {
					return nil, "Please provide environment to auto-complete application names. '--environment <environment>' flag."
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

func environmentsFromCLIArgs() []string {
	seen := map[string]struct{}{}
	environments := make([]string, 0)
	args := os.Args

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-e" || arg == "--environment":
			if i+1 >= len(args) {
				continue
			}
			next := args[i+1]
			if strings.HasPrefix(next, "-") || next == "" {
				continue
			}
			if _, ok := seen[next]; !ok {
				seen[next] = struct{}{}
				environments = append(environments, next)
			}
			i++
		case strings.HasPrefix(arg, "--environment="):
			env := strings.TrimPrefix(arg, "--environment=")
			if env != "" {
				if _, ok := seen[env]; !ok {
					seen[env] = struct{}{}
					environments = append(environments, env)
				}
			}
		case strings.HasPrefix(arg, "-e="):
			env := strings.TrimPrefix(arg, "-e=")
			if env != "" {
				if _, ok := seen[env]; !ok {
					seen[env] = struct{}{}
					environments = append(environments, env)
				}
			}
		}
	}

	return environments
}
