package command

import (
	"context"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func activity(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Activity{
		Job:    parentFlags,
		Output: "table",
		Limit:  20,
	}

	return &naistrix.Command{
		Name:        "activity",
		Title:       "Show activity for a job.",
		Description: "Displays recent events for a specific job, such as triggers, completions, and failures.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.GetJobActivity(ctx, flags.Team, args.Get("name"), string(flags.Environment), flags.Limit)
			if err != nil {
				return err
			}
			if len(ret) == 0 {
				out.Println("No activity found for job.")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
		AutoCompleteFunc: autoCompleteJobNames(parentFlags),
	}
}
