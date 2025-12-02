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
		Title:       "Workload and team logs.",
		Description: "Fetch and stream logs from workloads and teams.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("exactly one environment must be specified")
			}
			if args.Get("name") == "" {
				return fmt.Errorf("application name is required")
			}

			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return fmt.Errorf("unable to get authenticated user: %w", err)
			}

			queryEnvironment := flags.Environment
			if user.Domain() == "nav.no" {
				queryEnvironment = flag.Env(strings.TrimSuffix(string(queryEnvironment), "-gcp"))
			}
			appName := args.Get("name")
			query := flags.RawQuery
			if query == "" {
				query = logs.NewQueryBuilder().
					AddEnvironments(string(queryEnvironment)).
					AddTeams(flags.Team).
					AddWorkloads(appName).
					AddContainers(flags.Container...).
					AddPods(flags.Instance...).
					Build()
			}

			if err := naisapi.TailLog(ctx, out, string(flags.Environment), flags.Limit, flags.Since, flags.WithTimestamps, flags.WithLabels, query); err != nil {
				return fmt.Errorf("unable to tail logs: %w", err)
			}

			return nil
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
