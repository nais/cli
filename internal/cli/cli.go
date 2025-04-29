package cli

import (
	"context"

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

func Run(ctx context.Context, args []string) error {
	app := &cli.Command{
		Name:                  "nais",
		Usage:                 "A Nais cli",
		Description:           "Nais platform utility cli, respects consoledonottrack.com",
		Version:               version + "-" + commit,
		EnableShellCompletion: true,
		HideHelpCommand:       true,
		Suggest:               true,
		Commands: []*cli.Command{
			gcpcommand.Login(),
			kubeconfigcommand.Kubeconfig(),
			validatecommand.Validate(),
			debugcommand.Debug(),
			aivencommand.Aiven(),
			naisdevicecommand.Device(),
			postgrescommand.Postgres(),
		},
	}

	metrics.CollectCommandHistogram(ctx, app.Commands)
	return app.Run(ctx, args)
}
