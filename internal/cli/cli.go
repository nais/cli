package cli

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	cobra.EnableTraverseRunHooks = true

	cmdFlags := &root.Flags{}
	cmd := &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}
	cmd.PersistentFlags().CountVarP(&cmdFlags.VerboseLevel, "verbose", "v", `Verbose output.
Use -v for info, -vv for debug, -vvv for trace.`)
	cmd.AddGroup(authGroup)
	cmd.AddCommand(
		loginCommand(cmdFlags),
		logoutCommand(cmdFlags),
		kubeconfigCommand(cmdFlags),
		validateCommand(cmdFlags),
		debugCommand(cmdFlags),
		aivenCommand(cmdFlags),
		deviceCommand(cmdFlags),
		postgresCommand(cmdFlags),
		alphaCommand(
			naisApiCommand(cmdFlags),
		),
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

	if err != nil {
		if errors.Is(err, naisapi.ErrNotAuthenticated) {
			// TODO(thokra): Auto login process of some kind
			// Check if interactive
			// fmt.Println("Please try to log in again. Press enter to start the login process, or Ctrl+C to cancel.")
			// Start login process, if successful, rerun the command
		}
	}

	return err
}
