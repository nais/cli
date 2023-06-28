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
	version = "dev"
	commit  = "none"
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
		Description:          "NAIS CLI",
		Version:              version + "-" + commit,
		EnableBashCompletion: true,
		HideHelpCommand:      true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
