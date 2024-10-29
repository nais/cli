package devicecmd

import (
	"github.com/nais/cli/pkg/metrics"
	"github.com/nais/cli/pkg/naisdevice"
	"github.com/urfave/cli/v2"
)

func connectCommand() *cli.Command {
	return &cli.Command{
		Name:  "connect",
		Usage: "Creates a naisdevice connection, will lock until connection",
		Action: func(context *cli.Context) error {
			metrics.AddOne("device", "device_connect_total")
			return naisdevice.Connect(context.Context)
		},
	}
}

func disconnectCommand() *cli.Command {
	return &cli.Command{
		Name:  "disconnect",
		Usage: "Disconnects your naisdevice",
		Action: func(context *cli.Context) error {
			metrics.AddOne("device", "device_disconnect_total")
			return naisdevice.Disconnect(context.Context)
		},
	}
}
