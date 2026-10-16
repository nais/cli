package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

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
