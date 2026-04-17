package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/flags"
	logflags "github.com/nais/cli/internal/log/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

func Log(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &logflags.LogFlags{
		GlobalFlags: parentFlags,
		Since:       time.Hour,
		Limit:       100,
	}
	return &naistrix.Command{
		Name:        "log",
		Aliases:     []string{"logs"},
		Title:       "Show logs for a team.",
		Description: "Fetch and stream logs from a team.",
		Flags:       flags,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("exactly one environment must be specified")
			}

			return nil
		},
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
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
					AddTeams(flags.Team).
					Build()
			}

			if err := naisapi.TailLog(ctx, out, flags.Environment, flags.Limit, flags.Since, flags.WithTimestamps, flags.WithLabels, query); err != nil {
				return fmt.Errorf("unable to tail logs: %w", err)
			}

			return nil
		},
	}
}
