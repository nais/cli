package command

import (
	"context"
	"fmt"

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
			if len(flags.Team) == 0 {
				return fmt.Errorf("--team is required")
			}

			return nil
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			query := NewQueryBuilder().
				AddTeams(flags.Team...).
				AddEnvironments(flags.Environment...).
				AddWorkloads(flags.Workload...).
				AddContainers(flags.Container...).
				Build()

			if err := naisapi.TailLog(ctx, out, query); err != nil {
				return fmt.Errorf("unable to tail logs: %w", err)
			}

			return nil
		},
	}
}
