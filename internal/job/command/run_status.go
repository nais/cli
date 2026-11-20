package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func runStatus(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.RunStatus{Job: parentFlags}
	return &naistrix.Command{
		Name:        "status",
		Title:       "Show status and instances for a job run.",
		Description: "Shows the outcome and observed instances for a job run. Pass a job name and optionally a run name. If the run name is omitted, the latest run in the API's ordering is shown.",
		// Naistrix doesn't support optional positional arguments. Use a single
		// repeatable argument and enforce the two-position limit ourselves.
		Args:                   []naistrix.Argument{{Name: "job", Repeatable: true}},
		DisableArgumentPrompts: true,
		Flags:                  flags,
		ValidateFlagsFunc:      validation.RequireEnvironmentFlag(flags),
		ValidateArgsFunc: func(_ context.Context, args *naistrix.Arguments) error {
			values := args.GetRepeatable("job")
			if len(values) > 2 || values[0] == "" || len(values) == 2 && values[1] == "" {
				return naistrix.Errorf("Expected a job name and optionally a run name")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			switch args.Len() {
			case 0:
				names, err := jobNameChoices(parentFlags)(ctx, args)
				if err != nil {
					return nil, err.Error()
				}
				return names, "Select a job."
			case 1:
				names, err := job.RecentRunNames(ctx, flags.Team, string(flags.Environment), args.GetRepeatable("job")[0])
				if err != nil {
					return nil, fmt.Sprintf("Unable to fetch job runs: %v", err)
				}
				return names, "Select a recent run, or omit it to inspect the latest run."
			}
			return nil, ""
		},
		Examples: []naistrix.CommandExample{
			{Description: "Inspect the latest run in dev.", Command: "my-job -e dev"},
			{Description: "Inspect a specific run in dev.", Command: "my-job my-job-20261015-120000 -e dev"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			values := args.GetRepeatable("job")
			var runName string
			if len(values) == 2 {
				runName = values[1]
			}
			status, err := job.GetRunStatus(ctx, flags.Team, string(flags.Environment), values[0], runName)
			if err != nil {
				return err
			}
			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(status)
			}
			return renderRunStatus(out, status, flags.Team)
		},
	}
}

func renderRunStatus(out *naistrix.OutputWriter, status *job.RunStatus, team string) error {
	out.Printf("%s / %s (%s): %s\n", status.Job, status.Run, status.Environment, status.State)
	if status.Message != "" && !strings.EqualFold(status.Message, string(status.State)) {
		out.Println(status.Message)
	}
	if len(status.Instances) == 0 {
		out.Println("No instances observed for this run.")
		out.Printf("Check workload history: nais job events %s -t %s -e %s\n", status.Job, team, status.Environment)
		return nil
	}

	type row struct {
		Name, State, Image string
		Restarts           int
		Created            string
	}
	rows := make([]row, 0, len(status.Instances))
	for _, inst := range status.Instances {
		label := string(inst.State)
		if inst.State == gql.JobRunInstanceStatePending {
			label = "Waiting for placement"
		}
		rows = append(rows, row{inst.Name, label, inst.Image, inst.Restarts, inst.Created.Format(time.RFC3339)})
	}
	if err := out.Table().Render(rows); err != nil {
		return err
	}
	for _, inst := range status.Instances {
		if inst.Message != "" {
			out.Printf("  %s: %s\n", inst.Name, inst.Message)
		}
		if inst.LastExit != nil {
			out.Printf("  %s last exit: %s\n", inst.Name, inst.LastExit)
		}
	}
	return nil
}
