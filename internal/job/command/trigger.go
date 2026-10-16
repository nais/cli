package command

import (
	"context"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func trigger(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Trigger{Job: parentFlags}

	return &naistrix.Command{
		Name:        "trigger",
		Title:       "Trigger a job manually.",
		Description: "Creates a new run of the specified job outside of its normal schedule. Requires exactly one environment to be specified.",
		Args: []naistrix.Argument{
			{Name: "name", Prompt: "Name of the job to trigger", ChoicesFunc: jobNameChoices(parentFlags)},
		},
		Flags:             flags,
		ValidateFlagsFunc: validation.RequireEnvironmentFlag(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.TriggerJob(ctx, flags.Team, args.Get("name"), string(flags.Environment), flags.RunName)
			if err != nil {
				return err
			}

			out.Println(ret)
			return nil
		},
	}
}
