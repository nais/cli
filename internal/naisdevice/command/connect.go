package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/naistrix"
)

func connectcmd(parentFlags *flag.Device) *naistrix.Command {
	flags := &flag.Connect{Device: parentFlags}
	return &naistrix.Command{
		Name:        "connect",
		Title:       "Connect your naisdevice.",
		Description: "Connect to the naisdevice VPN. This establishes a connection to all non-privileged gateways automatically. By default, waits until all non-privileged (non-JITA) gateways are healthy; use --no-wait to return as soon as the connection is established.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return naisdevice.Connect(ctx, !flags.NoWait, out)
		},
	}
}
