package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func connect() *cli.Command {
	return &cli.Command{
		Name:  "connect",
		Usage: "Creates a naisdevice connection, will lock until connection",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return naisdevice.Connect(ctx)
		},
	}
}

func disconnect() *cli.Command {
	return &cli.Command{
		Name:  "disconnect",
		Usage: "Disconnects your naisdevice",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return naisdevice.Disconnect(ctx)
		},
	}
}
