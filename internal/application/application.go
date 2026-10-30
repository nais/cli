package application

import (
	"context"
	"io"
	"os"
	"slices"

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
	globalFlags := &flags.GlobalFlags{AdditionalFlags: &flags.AdditionalFlags{}}
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

	return err
}
