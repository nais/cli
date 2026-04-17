package application

import (
	"context"
	"errors"
	"io"
	"os"
	"slices"

	activity "github.com/nais/cli/internal/activity/command"
	alpha "github.com/nais/cli/internal/alpha/command"
	appCommand "github.com/nais/cli/internal/app/command"
	"github.com/nais/cli/internal/auth"
	configCmd "github.com/nais/cli/internal/config/command"
	debug "github.com/nais/cli/internal/debug/command"
	"github.com/nais/cli/internal/flags"
	issues "github.com/nais/cli/internal/issues/command"
	jobCommand "github.com/nais/cli/internal/job/command"
	kafkaCommand "github.com/nais/cli/internal/kafka/command"
	kubeconfig "github.com/nais/cli/internal/kubeconfig/command"
	log "github.com/nais/cli/internal/log/command"
	members "github.com/nais/cli/internal/member/command"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/naisapi"
	naisapiauth "github.com/nais/cli/internal/naisapi/auth"
	naisapiCommand "github.com/nais/cli/internal/naisapi/command"
	naisdevice "github.com/nais/cli/internal/naisdevice/command"
	opensearchCommand "github.com/nais/cli/internal/opensearch/command"
	postgres "github.com/nais/cli/internal/postgres/command"
	secrets "github.com/nais/cli/internal/secret/command"
	status "github.com/nais/cli/internal/status/command"
	validate "github.com/nais/cli/internal/validate/command"
	valkeyCommand "github.com/nais/cli/internal/valkey/command"
	"github.com/nais/cli/internal/version"
	vulnerabilities "github.com/nais/cli/internal/vulnerability/command"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type Application struct {
	*naistrix.Application
	Commands []*naistrix.Command
}

func New(w io.Writer) (*Application, *flags.GlobalFlags, error) {
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

	naisapiauth.ConfigFilePath = &f.Config

	cmds := []*naistrix.Command{
		auth.Auth(globalFlags),
		activity.Activity(globalFlags),
		appCommand.App(globalFlags),
		jobCommand.Job(globalFlags),
		kafkaCommand.Kafka(globalFlags),
		opensearchCommand.OpenSearch(globalFlags),
		valkeyCommand.Valkey(globalFlags),
		naisdevice.Naisdevice(globalFlags),
		members.Members(globalFlags),
		alpha.Alpha(globalFlags),
		postgres.Postgres(globalFlags),
		debug.Debug(globalFlags),
		kubeconfig.Kubeconfig(globalFlags),
		configCmd.Config(globalFlags),
		secrets.Secrets(globalFlags),
		vulnerabilities.Vulnerabilities(globalFlags),
		validate.Validate(globalFlags),
		issues.Issues(globalFlags),
		log.Log(globalFlags),
		status.Status(globalFlags),
		naisapiCommand.Api(globalFlags),
	}

	if err = app.AddCommand(cmds[0], cmds[1:]...); err != nil {
		return nil, nil, err
	}

	return &Application{Application: app, Commands: cmds}, globalFlags, nil
}

func Run(ctx context.Context, w io.Writer) error {
	app, f, err := New(w)
	if err != nil {
		return err
	}

	err = app.Run(naistrix.RunWithContext(ctx))
	autoComplete := slices.Contains(os.Args[1:], "__complete")

	if !autoComplete {
		flushMetrics := metric.Initialize(f.IsTrace())
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
		if errors.Is(err, naisapi.ErrNeedsLogin) {
			pterm.Println()
			pterm.Warning.Println("You must (re-)authenticate to run this command.")

			if !autoComplete && term.IsTerminal(int(os.Stdin.Fd())) { // #nosec G115
				pterm.Println()
				result, _ := pterm.DefaultInteractiveConfirm.
					WithDefaultValue(true).
					Show("Would you like to log in and re-run the command?")

				if result {
					pterm.Println()
					if err := naisapi.Login(ctx, naistrix.NewOutputWriter(w, &f.VerboseLevel)); err != nil {
						return err
					}

					pterm.Println()
					pterm.Info.Printf("Re-running command %s\n", executedCommand)

					pterm.Println()
					return app.Run(naistrix.RunWithContext(ctx))
				}
			}

			pterm.Println()
			pterm.Error.Println("Please run `nais login --nais` to (re-)authenticate.")
			return nil
		}

		return err
	}

	return nil
}
