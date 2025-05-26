package naisdevice

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice/config"
	"github.com/nais/cli/internal/naisdevice/connect"
	"github.com/nais/cli/internal/naisdevice/disconnect"
	"github.com/nais/cli/internal/naisdevice/doctor"
	"github.com/nais/cli/internal/naisdevice/jita"
	"github.com/nais/cli/internal/naisdevice/status"
	"github.com/nais/cli/internal/root"
)

func Command(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("device", "Interact with naisdevice.", cli.WithSubCommands(
		status.Command(rootFlags),
		jita.Command(rootFlags),
		doctor.Command(rootFlags),
		disconnect.Command(rootFlags),
		connect.Command(rootFlags),
		config.Command(rootFlags),
	))
}
