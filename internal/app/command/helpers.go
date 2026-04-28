package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

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
