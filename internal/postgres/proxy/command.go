package proxy

import (
	"context"

	"github.com/nais/cli/internal/postgres"
)

type Flags struct {
	*postgres.Flags
	Port uint
	Host string
}

func Run(ctx context.Context, applicationName string, flags *Flags) error {
	return postgres.RunProxy(ctx, applicationName, flags.Context, flags.Namespace, flags.Host, flags.Port, flags.IsVerbose())
}
