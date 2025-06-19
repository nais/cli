package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
)

func connectcmd(_ *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "connect",
		Title: "Connect your naisdevice.",
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			return naisdevice.Connect(ctx, out)
		},
	}
}
