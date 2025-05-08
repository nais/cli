package cli

import (
	"context"
	"strconv"

	"github.com/nais/cli/internal/root"
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
	app.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

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

func parseRootFlags(cmd *cobra.Command, flags *root.Flags) error {
	if verbose, err := strconv.ParseBool(cmd.Flag("verbose").Value.String()); err != nil {
		return err
	} else {
		flags.Verbose = verbose
	}

	return nil
}
