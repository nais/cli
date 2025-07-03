package command

import (
	"context"

	"github.com/nais/cli/v2/internal/naisdevice"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func connectcmd(_ *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "connect",
		Title: "Connect your naisdevice.",
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			return naisdevice.Connect(ctx, out)
		},
	}
}
