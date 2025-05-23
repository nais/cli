package naisdevice

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice/status"
	"github.com/nais/cli/internal/root"
)

func Device(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("device", "Interact with naisdevice.", cli.WithSubCommands(
		status.Status(rootFlags),
	))
}
