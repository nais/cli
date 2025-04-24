package command

import (
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v2"
)

func connect() *cli.Command {
	return &cli.Command{
		Name:  "connect",
		Usage: "Creates a naisdevice connection, will lock until connection",
		Action: func(context *cli.Context) error {
			return naisdevice.Connect(context.Context)
		},
	}
}

func disconnect() *cli.Command {
	return &cli.Command{
		Name:  "disconnect",
		Usage: "Disconnects your naisdevice",
		Action: func(context *cli.Context) error {
			return naisdevice.Disconnect(context.Context)
		},
	}
}
