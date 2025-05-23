package cli

import (
	"github.com/spf13/cobra"
)

func alphaCommand(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Alpha versions of Nais CLI commands.",
	}
	cmd.AddCommand(cmds...)
	return cmd
}
