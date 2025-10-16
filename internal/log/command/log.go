package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/alpha/command/flag"
	logflags "github.com/nais/cli/internal/log/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

func Log(parentFlags *flag.Alpha) *naistrix.Command {
	flags := &logflags.LogFlags{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "log",
		Title:       "Workload and team logs.",
		Description: "Fetch and stream logs from workloads and teams.",
		Flags:       flags,
		ValidateFunc: func(ctx context.Context, args []string) error {
			if flags.Environment == "" {
				return fmt.Errorf("--environment is required")
			}

			if len(flags.Team) == 0 {
				return fmt.Errorf("--team is required")
			}

			return nil
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return fmt.Errorf("unable to get authenticated user: %w", err)
			}

			queryEnvironment := flags.Environment
			if user.Domain() == "nav.no" {
				queryEnvironment = strings.TrimSuffix(queryEnvironment, "-gcp")
			}

			query := flags.RawQuery
			if query == "" {
				query = NewQueryBuilder().
					AddEnvironments(queryEnvironment).
					AddTeams(flags.Team...).
					AddWorkloads(flags.Workload...).
					AddContainers(flags.Container...).
					Build()
			}

			if err := naisapi.TailLog(ctx, out, flags.Environment, query, flags.WithTimestamps, flags.WithLabels); err != nil {
				return fmt.Errorf("unable to tail logs: %w", err)
			}

			return nil
		},
	}
}
