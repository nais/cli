package flag

import (
	"github.com/nais/cli/internal/root"
)

type Application struct {
	*root.Flags
}

type Create struct {
	*Application
	Name string `name:"name" short:"a" usage:"Name of the application to initialize."`
	Team string `name:"team" short:"t" usage:"Name of the team who owns the application."`
}
