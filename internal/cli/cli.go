package cli

import (
	"github.com/spf13/cobra"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func Run() error {
	app := &cobra.Command{
		Use:     "nais",
		Short:   "A Nais cli",
		Long:    "Nais platform utility cli, respects consoledonottrack.com",
		Version: version + "-" + commit,
	}

	app.AddCommand(
		login(),
		kubeconfig(),
		validate(),
		debug(),
		aiven(),
		device(),
		postgres(),
	)

	return app.Execute()
}
