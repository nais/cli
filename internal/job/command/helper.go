package command

import (
	"context"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func autoCompleteJobNames(flags *flag.Job) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != 0 {
			return nil, ""
		}

		if len(flags.Team) == 0 {
			return nil, "Please provide team to auto-complete job names. 'nais defaults set team <team>', or '--team <team>' flag."
		}

		if flags.Environment == "" {
			return nil, "Please provide environment to auto-complete job names. '-e, --environment <environment>' flag."
		}

		jobs, err := job.GetJobNames(ctx, flags.Team, string(flags.Environment))
		if err != nil {
			return nil, "Unable to fetch job names."
		}

		return jobs, "Select a job."
	}
}
