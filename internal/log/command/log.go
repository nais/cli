package command

import (
	"context"

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
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return naisapi.TailLog(ctx, out, flags)
		},
	}
}
