package add

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

type Arguments struct {
	ApplicationName string
	Username        string
	Password        string
}

func Run(ctx context.Context, args Arguments, flags *flag.UserAdd) error {
	return postgres.AddUser(ctx, args.ApplicationName, args.Username, args.Username, flags.Context, flags.Namespace, flags.Privilege)
}
