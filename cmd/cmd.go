package cmd

import (
	"log"
	"os"

	"github.com/nais/cli/cmd/aivencmd"
	"github.com/nais/cli/cmd/appstartercmd"
	"github.com/nais/cli/cmd/devicecmd"
	"github.com/nais/cli/cmd/kubeconfigcmd"
	"github.com/nais/cli/cmd/postgrescmd"
	"github.com/nais/cli/cmd/rootcmd"
	"github.com/nais/cli/cmd/validatecmd"
	"github.com/urfave/cli/v2"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func commands() []*cli.Command {
	return append(
		rootcmd.Commands(),
		aivencmd.Command(),
		appstartercmd.Command(),
		devicecmd.Command(),
		kubeconfigcmd.Command(),
		postgrescmd.Command(),
		validatecmd.Command(),
	)
}

func Run() {
	app := &cli.App{
		Name:                 "nais",
		Usage:                "A NAIS CLI",
		Description:          "A simple CLI application that developers in NAV can use",
		Version:              version + "-" + commit,
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Suggest:              true,
		Commands:             commands(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
