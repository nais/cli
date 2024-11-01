package cmd

import (
	"log"
	"os"

	"github.com/nais/cli/cmd/aivencmd"
	"github.com/nais/cli/cmd/devicecmd"
	"github.com/nais/cli/cmd/kubeconfigcmd"
	"github.com/nais/cli/cmd/postgrescmd"
	"github.com/nais/cli/cmd/rootcmd"
	"github.com/nais/cli/cmd/validatecmd"
	m "github.com/nais/cli/pkg/metrics"
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
		devicecmd.Command(),
		kubeconfigcmd.Command(),
		postgrescmd.Command(),
		validatecmd.Command(),
	)
}

func Run() {
	app := &cli.App{
		Name:                 "nais",
		Usage:                "A Nais cli",
		Description:          "Nais platform utility cli, respects consoledonottrack.com",
		Version:              version + "-" + commit,
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Suggest:              true,
		Commands:             commands(),
	}

	m.CollectCommandHistogram(app.Commands)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
