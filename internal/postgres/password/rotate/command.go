package rotate

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.PasswordRotate) error {
	return postgres.RotatePassword(ctx, applicationName, flags.Context, flags.Namespace)
}
