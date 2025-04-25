package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	aivencommand "github.com/nais/cli/internal/aiven/command"
	debugcommand "github.com/nais/cli/internal/debug/command"
	gcpcommand "github.com/nais/cli/internal/gcp/command"
	kubeconfigcommand "github.com/nais/cli/internal/kubeconfig/command"
	"github.com/nais/cli/internal/metrics"
	naisdevicecommand "github.com/nais/cli/internal/naisdevice/command"
	postgrescommand "github.com/nais/cli/internal/postgres/command"
	validatecommand "github.com/nais/cli/internal/validate/command"
	"github.com/urfave/cli/v3"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func commands() []*cli.Command {
	return append(
		[]*cli.Command{gcpcommand.Login()},
		kubeconfigcommand.Kubeconfig(),
		validatecommand.Validate(),
		debugcommand.Debug(),
		aivencommand.Aiven(),
		naisdevicecommand.Device(),
		postgrescommand.Postgres(),
	)
}

func Run(ctx context.Context) {
	app := &cli.Command{
		Name:                  "nais",
		Usage:                 "A Nais cli",
		Description:           "Nais platform utility cli, respects consoledonottrack.com",
		Version:               version + "-" + commit,
		EnableShellCompletion: true,
		HideHelpCommand:       true,
		Suggest:               true,
		Commands:              commands(),
	}

	metrics.CollectCommandHistogram(ctx, app.Commands)

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

	err := app.Run(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
