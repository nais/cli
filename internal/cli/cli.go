package cli

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	cmdFlags := root.Flags{}
	cmd := &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}
	cmd.PersistentFlags().CountVarP(&cmdFlags.VerboseLevel, "verbose", "v", "Verbose output.")
	cmd.AddCommand(
		login(&cmdFlags),
		kubeconfig(&cmdFlags),
		validate(&cmdFlags),
		debug(&cmdFlags),
		aiven(&cmdFlags),
		device(&cmdFlags),
		postgres(&cmdFlags),
	)

	fmt.Printf("OnInitialize, verbose: %v\n", cmdFlags.VerboseLevel)
	flushMetrics := metric.Initialize()

	executedCommand, err := cmd.ExecuteContextC(ctx)
	if executedCommand != nil {
		collectCommandHistogram(ctx, cmd, err)
	}

	// This has to be called _after_ cmd.ExecuteContextC, as we don't know the verbosity level before that
	flushMetrics(cmdFlags.IsDebug())

	return err
}
