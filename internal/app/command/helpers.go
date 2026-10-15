package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func autoCompleteAppNames(flags *flag.App) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != 0 {
			return nil, ""
		}

		apps, err := app.GetApplicationNames(ctx, flags.Team)
		if err != nil {
			return nil, "Unable to fetch application names."
		}

		if flags.Environment != "" {
			return apps.InEnv(string(flags.Environment)), "Select an application."
		}

		return apps.Unique(), "Select an application. Use --environment to filter by environment."
	}
}
