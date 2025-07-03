package application

import (
	"context"
	"io"
	"os"
	"slices"

	aiven "github.com/nais/cli/v2/internal/aiven/command"
	alpha "github.com/nais/cli/v2/internal/alpha/command"
	login "github.com/nais/cli/v2/internal/auth/login"
	logout "github.com/nais/cli/v2/internal/auth/logout"
	debug "github.com/nais/cli/v2/internal/debug/command"
	kubeconfig "github.com/nais/cli/v2/internal/kubeconfig/command"
	"github.com/nais/cli/v2/internal/metric"
	naisdevice "github.com/nais/cli/v2/internal/naisdevice/command"
	postgres "github.com/nais/cli/v2/internal/postgres/command"
	"github.com/nais/cli/v2/internal/root"
	validate "github.com/nais/cli/v2/internal/validate/command"
	"github.com/nais/cli/v2/internal/version"
	"github.com/nais/naistrix"
)

func newApplication(flags *root.Flags) *naistrix.Application {
	return &naistrix.Application{
		Name:    "nais",
		Title:   "Nais CLI",
		Version: version.Version,
		SubCommands: []*naistrix.Command{
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
	executedCommand, err := app.Run(ctx, naistrix.NewWriter(w), os.Args[1:])
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
