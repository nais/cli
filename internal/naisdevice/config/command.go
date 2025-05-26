package config

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice/config/get"
	"github.com/nais/cli/internal/naisdevice/config/set"
	"github.com/nais/cli/internal/root"
)

func Command(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("config", "Adjust or view the naisdevice configuration.", cli.WithSubCommands(
		set.Set(rootFlags),
		get.Get(rootFlags),
	))
}
