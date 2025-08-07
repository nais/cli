package flag

import (
	"github.com/nais/cli/internal/root"
)

type Init struct {
	*root.Flags
	Application string `name:"application" short:"a" usage:"Name of the application to initialize."`
	Team        string `name:"team" short:"t" usage:"Name of the team who owns the application."`
}
