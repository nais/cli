package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	logs "github.com/nais/cli/internal/log/command"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

func log(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Log{
		App:   parentFlags,
		Since: time.Hour,
		Limit: 100,
	}
	return &naistrix.Command{
		Name:        "log",
		Aliases:     []string{"logs"},
		Title:       "Show logs for an application.",
		Description: "Fetch and stream logs from an application.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return fmt.Errorf("unable to get authenticated user: %w", err)
			}

			appName := args.Get("name")
			environment, err := resolveAppEnvironment(ctx, out, flags.Team, appName, string(flags.Environment), false)
			if err != nil {
				return err
			}

			queryEnvironment := environment
			if user.Domain() == "nav.no" {
				queryEnvironment = strings.TrimSuffix(environment, "-gcp")
			}
			query := flags.RawQuery
			if query == "" {
				query = logs.NewQueryBuilder().
					AddEnvironments(queryEnvironment).
					AddTeams(flags.Team).
					AddWorkloads(appName).
					AddContainers(flags.Container...).
					AddPods(flags.Instance...).
					Build()
			}

			if err := naisapi.TailLog(ctx, out, environment, flags.Limit, flags.Since, flags.WithTimestamps, flags.WithLabels, query); err != nil {
				return fmt.Errorf("unable to tail logs: %w", err)
			}

			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if len(flags.Team) == 0 {
					return nil, "Please provide team to auto-complete application names. 'nais defaults set team <team>', or '--team <team>' flag."
				}

				envs := []string{}
				if flags.Environment != "" {
					envs = []string{string(flags.Environment)}
				}
				apps, err := app.GetApplicationNames(ctx, flags.Team, envs)
				if err != nil {
					return nil, "Unable to fetch application names."
				}
				return apps, "Select an application."
			}
			return nil, ""
		},
	}
}
