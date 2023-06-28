package deviceCmd

import (
	"github.com/nais/cli/pkg/naisdevice"
	"github.com/urfave/cli/v2"
)

func connectCommand() *cli.Command {
	return &cli.Command{
		Name:  "connect",
		Usage: "Creates a naisdevice connection, will lock until connection",
		Action: func(context *cli.Context) error {
			return naisdevice.Connect(context.Context)
		},
	}
}

func disconnectCommand() *cli.Command {
	return &cli.Command{
		Name:  "disconnect",
		Usage: "Disconnects your naisdevice",
		Action: func(context *cli.Context) error {
			return naisdevice.Disconnect(context.Context)
		},
	}
}
