package command

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/naistrix"
)

func restart(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Restart{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "restart",
		Title: "Restart an application.",
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() != 0 {
				return nil, ""
			}
			if len(flags.Team) == 0 {
				return nil, "Please provide team to auto-complete application names. 'nais config set team <team>', or '--team <team>' flag."
			}
			envs := []string(flags.Environment)
			if len(envs) != 1 {
				envs = cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
			}
			if len(envs) != 1 {
				return nil, "Please provide exactly one environment to auto-complete application names. '--environment <environment>' flag."
			}

			apps, err := app.GetApplicationNames(ctx, flags.Team, envs)
			if err != nil {
				return nil, fmt.Sprintf("Unable to fetch application names: %v", err)
			}
			return apps, "Select an application."
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := app.RestartApp(ctx, flags.Team, args.Get("name"), flags.Environment)
			if err != nil {
				return err
			}

			out.Println(ret)
			return nil
		},
	}
}
