package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
)

func disconnectcmd(_ *root.Flags) *cli.Command {
	return &cli.Command{
		Name: "disconnect", Short: "Disconnect your naisdevice.",
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			return naisdevice.Disconnect(ctx, out)
		},
	}
}
