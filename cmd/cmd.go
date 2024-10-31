package cmd

import (
	"context"
	"log"
	"os"

	"github.com/nais/cli/cmd/aivencmd"
	"github.com/nais/cli/cmd/appstartercmd"
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
		appstartercmd.Command(),
		devicecmd.Command(),
		kubeconfigcmd.Command(),
		postgrescmd.Command(),
		validatecmd.Command(),
	)
}

func collectCommandHistogram(app *cli.App) {
	ctx := context.Background()
	var validSubcommands []string
	for _, command := range app.Commands {
		validSubcommands = append(validSubcommands, command.Name)
		for _, subcommand := range command.Subcommands {
			validSubcommands = append(validSubcommands, subcommand.Name)
		}
	}

	doNotTrack := os.Getenv("DO_NOT_TRACK")
	if doNotTrack == "1" {
		log.Default().Println("DO_NOT_TRACK is set, not collecting metrics")
	}

	provider := m.NewMeterProvider()
	defer provider.Shutdown(ctx)
	// Record usages of subcommands that are exactly in the list of args we have, nothing else
	m.RecordCommandUsage(ctx, provider, m.Intersection(os.Args, validSubcommands))
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

	collectCommandHistogram(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
