package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func autoCompleteAppNames(flags *flag.App) naistrix.AutoCompleteFunc {
	choices := appNameChoices(flags)
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != 0 {
			return nil, ""
		}

		names, err := choices(ctx, args)
		if err != nil {
			return nil, "Unable to fetch application names."
		}

		if flags.Environment != "" {
			return names, "Select an application."
		}

		return names, "Select an application. Use --environment to filter by environment."
	}
}

func appNameChoices(flags *flag.App) naistrix.ArgumentChoicesFunc {
	return func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
		apps, err := app.GetApplicationNames(ctx, flags.Team)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch application names: %w", err)
		}
		if flags.Environment != "" {
			return apps.InEnv(string(flags.Environment)), nil
		}
		return apps.Unique(), nil
	}
}
