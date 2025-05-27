package proxy

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.Proxy) error {
	return postgres.RunProxy(ctx, applicationName, flags.Context, flags.Namespace, flags.Host, flags.Port, flags.IsVerbose())
}
