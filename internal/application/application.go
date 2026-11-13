package application

import (
	"context"
	"io"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	activityCommand "github.com/nais/cli/internal/activity/command"
	alphaCommand "github.com/nais/cli/internal/alpha/command"
	appCommand "github.com/nais/cli/internal/app/command"
	applyCommand "github.com/nais/cli/internal/apply/command"
	"github.com/nais/cli/internal/auth"
	authCommand "github.com/nais/cli/internal/auth/command"
	cloudsqlCommand "github.com/nais/cli/internal/cloudsql/command"
	configCommand "github.com/nais/cli/internal/config/command"
	debugCommand "github.com/nais/cli/internal/debug/command"
	"github.com/nais/cli/internal/flags"
	issuesCommand "github.com/nais/cli/internal/issues/command"
	jobCommand "github.com/nais/cli/internal/job/command"
	kafkaCommand "github.com/nais/cli/internal/kafka/command"
	kubeconfigCommand "github.com/nais/cli/internal/kubeconfig/command"
	logCommand "github.com/nais/cli/internal/log/command"
	memberCommand "github.com/nais/cli/internal/member/command"
	"github.com/nais/cli/internal/metric"
	naisapiauth "github.com/nais/cli/internal/naisapi/auth"
	naisapiCommand "github.com/nais/cli/internal/naisapi/command"
	naisdeviceCommand "github.com/nais/cli/internal/naisdevice/command"
	opensearchCommand "github.com/nais/cli/internal/opensearch/command"
	postgresCommand "github.com/nais/cli/internal/postgres/command"
	secretCommand "github.com/nais/cli/internal/secret/command"
	statusCommand "github.com/nais/cli/internal/status/command"
	validateCommand "github.com/nais/cli/internal/validate/command"
	valkeyCommand "github.com/nais/cli/internal/valkey/command"
	"github.com/nais/cli/internal/version"
	vulnerabilityCommand "github.com/nais/cli/internal/vulnerability/command"
	"github.com/nais/naistrix"
)

type Application struct {
	*naistrix.Application
	Commands []*naistrix.Command
}

func New(w io.Writer) (*Application, *flags.GlobalFlags, error) {
	return newApplication(w, auth.NaisLoginHook)
}

func newApplication(w io.Writer, hookFactories ...func(*flags.GlobalFlags) naistrix.PreRunHook) (*Application, *flags.GlobalFlags, error) {
	globalFlags := &flags.GlobalFlags{AdditionalFlags: &flags.AdditionalFlags{ConfirmLogin: true}}
	var hooks []naistrix.PreRunHook
	for _, factory := range hookFactories {
		hooks = append(hooks, factory(globalFlags))
	}
	app, f, err := naistrix.NewApplication(
		"nais",
		"Nais CLI",
		version.Version,
		naistrix.ApplicationWithWriter(w),
		naistrix.ApplicationWithPreRunHooks(hooks...),
		naistrix.ApplicationWithDefaultsCommandSkipPreRunHooks(auth.NaisLoginHookName),
		naistrix.ApplicationWithDefaultsDocumentation(
			naistrix.DefaultDefinition{
				Key:   "confirm-login",
				Title: "Ask before starting required Nais or Google Cloud login flows.",
				Description: heredoc.Doc(`
					Defaults to <info>true</info>. When a command requires a Nais or Google Cloud login and you are using an interactive terminal, the CLI asks for confirmation before starting the required login flow. If you are already logged in, no confirmation is needed.

					Set <info>nais defaults set confirm-login false</info> to start required login flows without this confirmation. This does not skip authentication or any browser interaction required by the login provider, and it does not disable other command confirmations.

					Override it for one command with <info>--confirm-login=false</info> to skip the confirmation, or <info>--confirm-login=true</info> to ask.

					Restore confirmations with <info>nais defaults set confirm-login true</info>, or remove the configured value with <info>nais defaults unset confirm-login</info> to use the built-in default.
				`),
			},
			naistrix.DefaultDefinition{
				Key:   "team",
				Title: "Default Nais team for team-owned resources.",
				Description: heredoc.Doc(`
					The Nais team whose applications, jobs, and other team-owned resources you want to work with. Use the team's name as shown in Nais Console.

					Set it with <info>nais defaults set team <team></info> to avoid repeating <info>--team</info> on commands that use a team.

					Override it for one command with <info>--team <team></info> (or <info>-t <team></info>), for example <info>nais app list --team <team></info>.

					Selecting a team does not grant access to its resources. Remove the default with <info>nais defaults unset team</info>.
				`),
			},
			naistrix.DefaultDefinition{
				Key:   "environment",
				Title: "Default Nais environment for resource selection and filtering.",
				Description: heredoc.Doc(`
					The Nais environment in which to select or filter resources.

					Set it with <info>nais defaults set environment <env></info> to avoid repeating <info>--environment</info> on commands that use an environment.

					Override it for one command with <info>--environment <env></info> (or <info>-e <env></info>), for example <info>nais app list --environment <env></info>.

					A stored environment also filters listings that would otherwise span environments, so check this default when resources seem to be missing.

					Remove it with <info>nais defaults unset environment</info>; behavior without an environment depends on the command.
				`),
			},
		),
	)
	if err != nil {
		return nil, nil, err
	}

	if err := app.AddGlobalFlags(globalFlags.AdditionalFlags); err != nil {
		return nil, nil, err
	}

	globalFlags.GlobalFlags = f

	naisapiauth.ConfigFilePath = &f.Config

	cmds := []*naistrix.Command{
		activityCommand.Activity(globalFlags),
		alphaCommand.Alpha(globalFlags),
		appCommand.App(globalFlags),
		applyCommand.Apply(globalFlags),
		authCommand.Auth(globalFlags),
		configCommand.Config(globalFlags),
		debugCommand.Debug(globalFlags),
		issuesCommand.Issues(globalFlags),
		jobCommand.Job(globalFlags),
		kafkaCommand.Kafka(globalFlags),
		kubeconfigCommand.Kubeconfig(globalFlags),
		logCommand.Log(globalFlags),
		memberCommand.Members(globalFlags),
		naisapiCommand.Api(globalFlags),
		naisdeviceCommand.Naisdevice(globalFlags),
		opensearchCommand.OpenSearch(globalFlags),
		postgresCommand.Postgres(globalFlags),
		cloudsqlCommand.CloudSQL(globalFlags),
		secretCommand.Secrets(globalFlags),
		statusCommand.Status(globalFlags),
		validateCommand.Validate(globalFlags),
		valkeyCommand.Valkey(globalFlags),
		vulnerabilityCommand.Vulnerabilities(globalFlags),
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

	// Includes interactive prompts and login flows, but not telemetry upload.
	start := time.Now()
	err = app.Run(naistrix.RunWithContext(ctx))
	duration := time.Since(start)
	autoComplete := isCompletionCommand(app.ExecutedCommand())

	if !autoComplete {
		flushMetrics := metric.Initialize()
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
			flushMetrics(f.IsTrace())
		}()
	}

	if !autoComplete {
		collectCommandMetrics(ctx, app.ExecutedCommand(), err, duration)
	}

	return err
}
