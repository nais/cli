package list

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.UserList) error {
	return postgres.ListUsers(ctx, applicationName, flags.Context, flags.Namespace)
}
