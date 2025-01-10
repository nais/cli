package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/nais/cli/cmd/debugcmd"

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
		debugcmd.Command(),
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

	// first, before running the cli propper we check if the argv[1] contains a
	// thing that is named nais-argv[1]. if so, we run that with the rest of the
	// argument string and then exit.
	// This gives us and our users a nice way of extending the cli by just shipping other
	// binaries. this is spiritually what git and others do.
	if len(os.Args) > 1 {
		if !isCommand(os.Args[1], app.Commands) {
			binaryName := "nais-" + os.Args[1]
			if err := runOtherBin(binaryName, os.Args[2:]); err == nil {
				os.Exit(0)

			}
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func isCommand(command string, commands []*cli.Command) bool {
	for _, cmd := range commands {
		if cmd.Name == command {
			return true
		}
	}
	return false
}

func runOtherBin(binary string, args []string) error {
	binaryPath, err := exec.LookPath(binary)
	if err != nil {
		return err
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
