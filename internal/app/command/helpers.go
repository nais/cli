package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

// requireSingleEnvironment validates that exactly one environment is specified.
func requireSingleEnvironment(envs flag.Environments) error {
	if len(envs) != 1 {
		return naistrix.Errorf("exactly one environment must be specified with -e/--environment")
	}
	return nil
}

// autoCompleteAppNames returns an AutoCompleteFunc that completes application names for the given flags.
func autoCompleteAppNames(flags *flag.App) func(context.Context, *naistrix.Arguments, string) ([]string, string) {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() == 0 {
			if len(flags.Team) == 0 {
				return nil, "Please provide team to auto-complete application names. 'nais defaults set team <team>', or '--team <team>' flag."
			}
			apps, err := app.GetApplicationNames(ctx, flags.Team, flags.Environment)
			if err != nil {
				return nil, "Unable to fetch application names."
			}
			return apps, "Select an application."
		}
		return nil, ""
	}
}
