package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/naisdevice"
	"github.com/nais/cli/v2/internal/root"
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
