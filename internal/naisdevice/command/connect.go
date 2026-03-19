package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/naistrix"
)

func connectcmd() *naistrix.Command {
	return &naistrix.Command{
		Name:        "connect",
		Title:       "Connect your naisdevice.",
		Description: "Connect to the naisdevice VPN. This establishes a connection to all non-privileged gateways automatically.",
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return naisdevice.Connect(ctx, out)
		},
	}
}
