package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func disconnectcmd(_ *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "disconnect",
		Title: "Disconnect your naisdevice.",
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			return naisdevice.Disconnect(ctx, out)
		},
	}
}
