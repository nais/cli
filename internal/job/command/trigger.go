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
		Name:  "trigger",
		Title: "Trigger a job manually.",
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
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if len(flags.Team) == 0 {
					return nil, "Please provide team to auto-complete job names. 'nais config set team <team>', or '--team <team>' flag."
				}
				if flags.Environment == "" {
					return nil, "Please provide environment to auto-complete job names. '--environment <environment>' flag."
				}
				jobs, err := job.GetJobNames(ctx, flags.Team, []string{string(flags.Environment)})
				if err != nil {
					return nil, "Unable to fetch job names."
				}
				return jobs, "Select a job."
			}
			return nil, ""
		},
	}
}
