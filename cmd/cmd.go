package cmd

import (
	"context"
	"log"
	"os"

	m "github.com/nais/cli/pkg/metrics"
	"go.opentelemetry.io/otel/metric"

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
		Usage:                "A Nais cli",
		Description:          "Nais platform utility cli, respects consoledonottrack.com",
		Version:              version + "-" + commit,
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Suggest:              true,
		Commands:             commands(),
	}

	var validSubcommands []string
	for _, command := range app.Commands {
		validSubcommands = append(validSubcommands, command.Name)
		for _, subcommand := range command.Subcommands {
			validSubcommands = append(validSubcommands, subcommand.Name)
		}
	}

	meterProv := m.New()
	defer meterProv.Shutdown(context.Background())

	commandHistogram, _ := meterProv.Meter("nais-cli").Int64Histogram("flag_usage", metric.WithDescription("Usage frequency of command flags"))

	// Record usages of subcommands that are exactly in the list of args we have, nothing else
	m.RecordCommandUsage(context.Background(), commandHistogram, m.Intersection(os.Args, validSubcommands))
	meterProv.ForceFlush(context.Background())

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
