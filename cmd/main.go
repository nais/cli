package main

import (
	"github.com/nais/cli/cmd/aivenCmd"
	"log"
	"os"

	"github.com/nais/cli/cmd/appStarterCmd"
	"github.com/nais/cli/cmd/deviceCmd"
	"github.com/nais/cli/cmd/validateCmd"
	"github.com/urfave/cli/v2"
)

var (
	// Is set during build
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "you"
)

func commands() []*cli.Command {
	return []*cli.Command{
		aivenCmd.Command(),
		appStarterCmd.Command(),
		deviceCmd.Command(),
		validateCmd.Command(),
	}
}

func main() {
	app := &cli.App{
		Name:                 "nais",
		Usage:                "NAIS CLI",
		Version:              version + "-" + commit,
		Description:          "NAIS Administrator CLI",
		Commands:             commands(),
		EnableBashCompletion: true,
		HideHelpCommand:      true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
