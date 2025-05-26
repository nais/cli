package disconnect

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice/naisdevicegrpc"
	"github.com/nais/cli/internal/root"
)

func Command(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("disconnect", "Disconnect your naisdevice.",
		cli.WithRun(run),
	)
}

func run(ctx context.Context, _ []string) error {
	return naisdevicegrpc.Disconnect(ctx)
}
