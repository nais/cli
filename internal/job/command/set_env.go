package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func setEnv(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.SetEnv{
		Job: parentFlags,
	}

	return &naistrix.Command{
		Name:        "env",
		Title:       "Set environment variables for a job.",
		Description: "Updates environment variables on the job. Use KEY=VALUE to set and KEY- to remove. Changes are temporary and will be overwritten on next deploy.",
		Examples: []naistrix.Example{
			{
				Description: "Set environment variables.",
				Command:     "my-job FOO=bar BAZ=qux -e dev",
			},
			{
				Description: "Remove an environment variable.",
				Command:     "my-job FOO- -e dev",
			},
			{
				Description: "Set and remove in one command.",
				Command:     "my-job NEW_VAR=hello OLD_VAR- -e dev",
			},
		},
		Args: []naistrix.Argument{
			{Name: "name"},
			{Name: "env_vars", Repeatable: true},
		},
		Flags: flags,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if flags.Environment == "" {
				return fmt.Errorf("environment must be specified (-e, --environment)")
			}
			if len(args.GetRepeatable("env_vars")) == 0 {
				return fmt.Errorf("at least one environment variable must be specified (KEY=VALUE or KEY-)")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")
			envVarArgs := args.GetRepeatable("env_vars")

			updates, err := job.ParseEnvVarUpdates(envVarArgs)
			if err != nil {
				return err
			}

			ret, err := job.SetJobEnv(ctx, flags.Team, name, string(flags.Environment), updates)
			if err != nil {
				return err
			}

			out.Println(ret)
			out.Warnf("Changes are temporary and will be overwritten on next deploy.\n")
			return nil
		},
		AutoCompleteFunc: autoCompleteJobNames(parentFlags),
	}
}
