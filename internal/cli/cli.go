package cli

import (
	"context"

	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

var (
	version = "local"
	commit  = "uncommited"
)

func Run(ctx context.Context) error {
	cmdFlags := root.Flags{}
	cmd := &cobra.Command{
		Use:          "nais",
		Long:         "Nais CLI",
		Version:      version + "-" + commit,
		SilenceUsage: true,
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

	return cmd.ExecuteContext(ctx)
}
