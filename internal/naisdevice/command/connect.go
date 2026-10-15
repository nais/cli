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
		Description: "Connect to the naisdevice VPN. Your device registers with the API server, which shares it with all gateways, and the command waits until every non-privileged (non-JITA) gateway responds as healthy.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return naisdevice.Connect(ctx, !flags.NoWait, out)
		},
	}
}
