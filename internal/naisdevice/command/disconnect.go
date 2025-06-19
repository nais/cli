package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
)

func disconnectcmd(_ *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "disconnect",
		Title: "Disconnect your naisdevice.",
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			return naisdevice.Disconnect(ctx, out)
		},
	}
}
