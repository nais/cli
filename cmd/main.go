package main

import (
	"github.com/nais/cli/cmd/aivenCmd"
	"github.com/nais/cli/cmd/kubeconfigCmd"
	"github.com/nais/cli/cmd/postgresCmd"
	"log"
	"os"

	"github.com/nais/cli/cmd/appStarterCmd"
	"github.com/nais/cli/cmd/deviceCmd"
	"github.com/nais/cli/cmd/validateCmd"
	"github.com/urfave/cli/v2"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func commands() []*cli.Command {
	return []*cli.Command{
		aivenCmd.Command(),
		appStarterCmd.Command(),
		deviceCmd.Command(),
		kubeconfigCmd.Command(),
		postgresCmd.Command(),
		validateCmd.Command(),
	}
}

func main() {
	app := &cli.App{
		Name:                 "nais",
		Usage:                "A NAIS CLI",
		Description:          "A simple CLI application that developers in NAV can use",
		Version:              version + "-" + commit,
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Commands:             commands(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
