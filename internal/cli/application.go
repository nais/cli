package cli

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

const (
	GroupAuthentication = "Authentication"
)

type Application struct {
	cobraCmd *cobra.Command
}

func NewApplication(flags *root.Flags, cmd ...*Command) *Application {
	cobra.EnableTraverseRunHooks = true

	cc := &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}

	setupFlags(flags, cc.PersistentFlags())

	cc.AddGroup(&cobra.Group{
		ID:    GroupAuthentication,
		Title: GroupAuthentication,
	})

	w := output.NewWriter(cc.OutOrStdout())

	for _, c := range cmd {
		c.init(w)
		cc.AddCommand(c.cobraCmd)
	}

	return &Application{
		cobraCmd: cc,
	}
}

func (a *Application) Run(ctx context.Context, flags *root.Flags) error {
	autoComplete := slices.Contains(os.Args[1:], "__complete")

	if !autoComplete {
		flushMetrics := metric.Initialize()
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
			flushMetrics(flags.IsDebug())
		}()
	}

	executedCommand, err := a.cobraCmd.ExecuteContextC(ctx)
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
		return err
	}

	return nil
}
