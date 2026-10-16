package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func jobNameChoices(flags *flag.Job) naistrix.ArgumentChoicesFunc {
	return func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
		fmt.Println(flags.Environment)
		jobs, err := job.GetJobNames(ctx, flags.Team, string(flags.Environment))
		if err != nil {
			return nil, fmt.Errorf("unable to fetch job names: %w", err)
		}
		return jobs, nil
	}
}
