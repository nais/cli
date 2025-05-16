package add

import (
	"context"

	"github.com/nais/cli/internal/postgres"
)

type Arguments struct {
	ApplicationName string
	Username        string
	Password        string
}

type Flags struct {
	*postgres.Flags
	Privilege string
}

func Run(ctx context.Context, args Arguments, flags *Flags) error {
	return postgres.AddUser(ctx, args.ApplicationName, args.Username, args.Username, flags.Context, flags.Namespace, flags.Privilege)
}
