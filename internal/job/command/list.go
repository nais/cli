package command

import (
	"context"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func list(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.List{Job: parentFlags}

	return &naistrix.Command{
		Name:  "list",
		Title: "List jobs in a team.",
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.GetTeamJobs(ctx, flags.Team, flags.Environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("Team has no jobs.")
				return nil
			}

			return out.Table().Render(ret)
		},
	}
}
