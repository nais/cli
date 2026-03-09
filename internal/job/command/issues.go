package command

import (
	"context"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func issues(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Issues{Job: parentFlags}

	return &naistrix.Command{
		Name:  "issues",
		Title: "Show issues for a job.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.GetJobIssues(ctx, flags.Team, args.Get("name"), flags.Environment)
			if err != nil {
				return err
			}
			if len(ret) == 0 {
				out.Println("No issues found for job.")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				jobs, err := job.GetJobNames(ctx, flags.Team)
				if err != nil {
					return nil, "Unable to fetch job names."
				}
				return jobs, "Select a job."
			}
			return nil, ""
		},
	}
}
