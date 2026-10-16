package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func autoCompleteJobNames(flags *flag.Job) naistrix.AutoCompleteFunc {
	choices := jobNameChoices(flags)
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != 0 {
			return nil, ""
		}

		if flags.Environment == "" {
			return nil, "Please provide environment to auto-complete job names. '-e, --environment <environment>' flag."
		}

		jobs, err := choices(ctx, args)
		if err != nil {
			return nil, "Unable to fetch job names."
		}

		return jobs, "Select a job."
	}
}

func jobNameChoices(flags *flag.Job) naistrix.ArgumentChoicesFunc {
	return func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
		jobs, err := job.GetJobNames(ctx, flags.Team, string(flags.Environment))
		if err != nil {
			return nil, fmt.Errorf("unable to fetch job names: %w", err)
		}
		return jobs, nil
	}
}
