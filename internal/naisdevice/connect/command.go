package connect

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice/naisdevicegrpc"
	"github.com/nais/cli/internal/root"
)

func Command(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("connect", "Connect your naisdevice.",
		cli.WithRun(run),
	)
}

func run(ctx context.Context, _ []string) error {
	return naisdevicegrpc.Connect(ctx)
}
