package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	globalFlags "github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	logs "github.com/nais/cli/internal/log/command"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

func log(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Log{
		Job:   parentFlags,
		Since: time.Hour,
		Limit: 100,
	}

	return &naistrix.Command{
		Name:        "log",
		Aliases:     []string{"logs"},
		Title:       "Show logs for a job.",
		Description: "Fetch and stream logs from a job.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("exactly one environment must be specified")
			}
			if args.Get("name") == "" {
				return fmt.Errorf("job name is required")
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
				queryEnvironment = globalFlags.Environment(strings.TrimSuffix(string(queryEnvironment), "-gcp"))
			}

			jobName := args.Get("name")
			query := flags.RawQuery
			if query == "" {
				query = logs.NewQueryBuilder().
					AddEnvironments(string(queryEnvironment)).
					AddTeams(flags.Team).
					AddWorkloads(jobName).
					AddContainers(flags.Container...).
					Build()
			}

			streamCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			var stoppedByTerminalState atomic.Bool
			go func() {
				ticker := time.NewTicker(3 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-streamCtx.Done():
						return
					case <-ticker.C:
						state, err := job.GetLatestJobRunState(streamCtx, flags.Team, jobName, string(flags.Environment))
						if err != nil {
							continue
						}
						if job.IsTerminalRunState(state) {
							stoppedByTerminalState.Store(true)
							cancel()
							return
						}
					}
				}
			}()

			if err := naisapi.TailLog(streamCtx, out, string(flags.Environment), flags.Limit, flags.Since, flags.WithTimestamps, flags.WithLabels, query); err != nil {
				if stoppedByTerminalState.Load() && errors.Is(err, context.Canceled) {
					return nil
				}
				return fmt.Errorf("unable to tail logs: %w", err)
			}

			return nil
		},
		AutoCompleteFunc: autoCompleteJobNames(parentFlags),
	}
}
