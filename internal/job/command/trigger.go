package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func trigger(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Trigger{Job: parentFlags}

	return &naistrix.Command{
		Name:        "trigger",
		Title:       "Trigger a job manually.",
		Description: "Creates a new run of the specified job outside of its normal schedule. Requires exactly one environment to be specified.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("exactly one environment must be specified")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.TriggerJob(ctx, flags.Team, args.Get("name"), string(flags.Environment), flags.RunName)
			if err != nil {
				return err
			}

			out.Println(ret)
			return nil
		},
		AutoCompleteFunc: autoCompleteJobNames(parentFlags),
	}
}
