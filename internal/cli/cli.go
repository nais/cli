package cli

import (
	"context"
	"os"
	"slices"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	cobra.EnableTraverseRunHooks = true

	cmdFlags := root.Flags{}
	cmd := &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}
	cmd.PersistentFlags().CountVarP(&cmdFlags.VerboseLevel, "verbose", "v", `Verbose output.
Use -v for info, -vv for debug, -vvv for trace.`)
	cmd.AddCommand(
		login(&cmdFlags),
		kubeconfig(&cmdFlags),
		validate(&cmdFlags),
		debug(&cmdFlags),
		aiven(&cmdFlags),
		device(&cmdFlags),
		postgres(&cmdFlags),
		apiProxy(&cmdFlags),
	)

	autoComplete := slices.Contains(os.Args[1:], "__complete")

	if !autoComplete {
		flushMetrics := metric.Initialize()
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
			flushMetrics(cmdFlags.IsDebug())
		}()
	}

	executedCommand, err := cmd.ExecuteContextC(ctx)
	if !autoComplete && executedCommand != nil {
		collectCommandHistogram(ctx, executedCommand, err)
	}

	return err
}
