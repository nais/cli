package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/naistrix"
)

func disconnectcmd() *naistrix.Command {
	return &naistrix.Command{
		Name:  "disconnect",
		Title: "Disconnect your naisdevice.",
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return naisdevice.Disconnect(ctx, out)
		},
	}
}
