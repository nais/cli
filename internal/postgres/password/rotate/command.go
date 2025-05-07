package rotate

import (
	"context"

	"github.com/nais/cli/internal/postgres"
)

func Run(ctx context.Context, applicationName string, flags postgres.Flags) error {
	return postgres.RotatePassword(ctx, applicationName, flags.Context, flags.Namespace)
}
