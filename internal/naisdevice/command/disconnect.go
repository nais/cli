package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func disconnectcmd(_ *root.Flags) *cli.Command {
	return cli.NewCommand("disconnect", "Disconnect your naisdevice.",
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			return naisdevice.Disconnect(ctx, out)
		}),
	)
}
