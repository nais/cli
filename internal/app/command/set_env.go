package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func setEnv(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.SetEnv{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "env",
		Title:       "Set environment variables for an application.",
		Description: "Updates environment variables on the application. Use KEY=VALUE to set and KEY- to remove. Changes are temporary and will be overwritten on next deploy.",
		Examples: []naistrix.Example{
			{
				Description: "Set environment variables.",
				Command:     "my-app FOO=bar BAZ=qux -e dev",
			},
			{
				Description: "Remove an environment variable.",
				Command:     "my-app FOO- -e dev",
			},
			{
				Description: "Set and remove in one command.",
				Command:     "my-app NEW_VAR=hello OLD_VAR- -e dev",
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

			updates, err := app.ParseEnvVarUpdates(envVarArgs)
			if err != nil {
				return err
			}

			ret, err := app.SetApplicationEnv(ctx, flags.Team, name, string(flags.Environment), updates)
			if err != nil {
				return err
			}

			out.Println(ret)
			out.Warnf("Changes are temporary and will be overwritten on next deploy.\n")
			return nil
		},
		AutoCompleteFunc: autoCompleteAppNames(parentFlags),
	}
}
