package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func connectcmd(_ *root.Flags) *cli.Command {
	return cli.NewCommand("connect", "Connect your naisdevice.",
		cli.WithRun(func(ctx context.Context, _ output.Output, _ []string) error {
			return naisdevice.Connect(ctx)
		}),
	)
}
