package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func deleteRun(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Delete{Job: parentFlags}

	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete a job run.",
		Description: "This command deletes an individual job run (a Kubernetes batch/v1 Job).",
		Args: []naistrix.Argument{
			{Name: "run-name"},
		},
		Flags: flags,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("exactly one environment must be specified")
			}
			if args.Get("run-name") == "" {
				return fmt.Errorf("run name is required")
			}
			return nil
		},
		Examples: []naistrix.Example{
			{
				Description: "Delete a job run named my-job-20250318-120000 in environment dev.",
				Command:     "my-job-20250318-120000 --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			runName := args.Get("run-name")

			if !flags.Yes {
				pterm.Warning.Printfln("You are about to delete job run %q in %q for team %q",
					runName, string(flags.Environment), flags.Team)

				result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
				if !result {
					return fmt.Errorf("cancelled by user")
				}
			}

			deleted, err := job.DeleteJobRun(ctx, flags.Team, string(flags.Environment), runName)
			if err != nil {
				return fmt.Errorf("deleting job run: %w", err)
			}

			if deleted {
				pterm.Success.Printfln("Deleted job run %q from %q for team %q",
					runName, string(flags.Environment), flags.Team)
			}

			return nil
		},
	}
}
