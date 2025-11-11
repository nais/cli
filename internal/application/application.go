package application

import (
	"context"
	"errors"
	"io"
	"os"
	"slices"

	aiven "github.com/nais/cli/internal/aiven/command"
	alpha "github.com/nais/cli/internal/alpha/command"
	login "github.com/nais/cli/internal/auth/login"
	logout "github.com/nais/cli/internal/auth/logout"
	debug "github.com/nais/cli/internal/debug/command"
	"github.com/nais/cli/internal/flags"
	kubeconfig "github.com/nais/cli/internal/kubeconfig/command"
	members "github.com/nais/cli/internal/member/command"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/naisapi"
	naisdevice "github.com/nais/cli/internal/naisdevice/command"
	postgres "github.com/nais/cli/internal/postgres/command"
	validate "github.com/nais/cli/internal/validate/command"
	"github.com/nais/cli/internal/version"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

type Application struct {
	*naistrix.Application
	Commands []*naistrix.Command
}

func newApplication(w io.Writer) (*Application, *flags.GlobalFlags, error) {
	app, f, err := naistrix.NewApplication(
		"nais",
		"Nais CLI",
		version.Version,
		naistrix.ApplicationWithWriter(w),
	)
	if err != nil {
		return nil, nil, err
	}

	additional := &flags.AdditionalFlags{}

	if err := app.AddGlobalFlags(additional); err != nil {
		return nil, nil, err
	}

	globalFlags := &flags.GlobalFlags{
		GlobalFlags:     f,
		AdditionalFlags: additional,
	}
	cmds := []*naistrix.Command{
		login.Login(globalFlags),
		logout.Logout(globalFlags),
		naisdevice.Naisdevice(globalFlags),
		members.Members(globalFlags),
		aiven.Aiven(globalFlags),
		alpha.Alpha(globalFlags),
		postgres.Postgres(globalFlags),
		debug.Debug(globalFlags),
		kubeconfig.Kubeconfig(globalFlags),
		validate.Validate(globalFlags),
	}

	if err = app.AddCommand(cmds[0], cmds[1:]...); err != nil {
		return nil, nil, err
	}

	return &Application{Application: app}, globalFlags, nil
}

func Run(ctx context.Context, w io.Writer) error {
	app, f, err := newApplication(w)
	if err != nil {
		return err
	}

	err = app.Run(naistrix.RunWithContext(ctx))
	autoComplete := slices.Contains(os.Args[1:], "__complete")

	if !autoComplete {
		flushMetrics := metric.Initialize()
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
			flushMetrics(f.IsTrace())
		}()
	}

	executedCommand := app.ExecutedCommand()
	if !autoComplete && executedCommand != nil {
		collectCommandHistogram(ctx, executedCommand, err)
	}

	if err != nil {
		if errors.Is(err, naisapi.ErrNotAuthenticated) {
			// TODO(tronghn): If tty; prompt for login (y/n)?
			pterm.Error.Println("You are not logged in. Please run `nais login -n` to authenticate.")
		}

		return err
	}

	return nil
}
