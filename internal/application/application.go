package application

import (
	"context"
	"io"
	"os"
	"slices"

	aiven "github.com/nais/cli/internal/aiven/command"
	alpha "github.com/nais/cli/internal/alpha/command"
	login "github.com/nais/cli/internal/auth/login"
	logout "github.com/nais/cli/internal/auth/logout"
	debug "github.com/nais/cli/internal/debug/command"
	kubeconfig "github.com/nais/cli/internal/kubeconfig/command"
	"github.com/nais/cli/internal/metric"
	naisdevice "github.com/nais/cli/internal/naisdevice/command"
	postgres "github.com/nais/cli/internal/postgres/command"
	"github.com/nais/cli/internal/root"
	validate "github.com/nais/cli/internal/validate/command"
	"github.com/nais/cli/internal/version"
	"github.com/nais/cli/pkg/cli"
)

func newApplication(flags *root.Flags) *cli.Application {
	return &cli.Application{
		Name:    "nais",
		Title:   "Nais CLI",
		Version: version.Version,
		SubCommands: []*cli.Command{
			login.Login(flags),
			logout.Logout(flags),
			naisdevice.Naisdevice(flags),
			aiven.Aiven(flags),
			alpha.Alpha(flags),
			postgres.Postgres(flags),
			debug.Debug(flags),
			kubeconfig.Kubeconfig(flags),
			validate.Validate(flags),
		},
		StickyFlags: flags,
	}
}

func Run(ctx context.Context, w io.Writer) error {
	flags := &root.Flags{}
	app := newApplication(flags)
	executedCommand, err := app.Run(ctx, cli.NewWriter(w), os.Args[1:])
	autoComplete := slices.Contains(os.Args[1:], "__complete")

	if !autoComplete {
		flushMetrics := metric.Initialize()
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
			flushMetrics(flags.IsTrace())
		}()
	}

	if !autoComplete && executedCommand != nil {
		collectCommandHistogram(ctx, executedCommand, err)
	}

	if err != nil {
		return err
	}

	return nil
}
