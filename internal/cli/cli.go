package cli

import (
	"context"

	"github.com/spf13/cobra"
)

var (
	version = "local"
	commit  = "uncommited"
)

func Run(ctx context.Context) error {
	app := &cobra.Command{
		Use:     "nais",
		Short:   "A Nais cli",
		Long:    "Nais platform utility cli, respects consoledonottrack.com",
		Version: version + "-" + commit,
	}
	fs := app.PersistentFlags()
	fs.BoolP("verbose", "v", false, "Verbose output")

	app.AddCommand(
		login(),
		kubeconfig(),
		validate(),
		debug(),
		aiven(),
		device(),
		postgres(),
	)

	return app.ExecuteContext(ctx)
}
