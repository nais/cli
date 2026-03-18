package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func listRuns(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.RunList{Job: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List runs for a job.",
		Description: "This command lists all runs for a specific job in a given environment.",
		Args: []naistrix.Argument{
			{Name: "job-name"},
		},
		Flags: flags,
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if flags.Team == "" {
					return nil, "Please provide team to auto-complete job names. 'nais config set team <team>', or '--team <team>' flag."
				}
				if flags.Environment == "" {
					return []string{"--environment"}, ""
				}
				jobs, err := job.GetJobNames(ctx, flags.Team, []string{string(flags.Environment)})
				if err != nil {
					return nil, "Unable to fetch job names."
				}
				return jobs, "Select a job."
			}
			return nil, ""
		},
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("exactly one environment must be specified")
			}
			if args.Get("job-name") == "" {
				return fmt.Errorf("job name is required")
			}
			return nil
		},
		Examples: []naistrix.Example{
			{
				Description: "List all runs for a job named my-job in the dev environment.",
				Command:     "my-job --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			jobName := args.Get("job-name")

			runs, err := job.GetJobRuns(ctx, flags.Team, string(flags.Environment), jobName)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(runs)
			}

			if len(runs) == 0 {
				out.Println("No runs found.")
				return nil
			}

			return out.Table().Render(runs)
		},
	}
}
