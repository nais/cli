package root

import (
	"fmt"
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version [flags]",
	Short: "Show 'nais-cli' client version",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := cmd.Flags().GetBool(helpers.CommitInformation)
		if err != nil {
			return err
		}

		if result {
			fmt.Printf("%s: %s commit: %s date: %s builtBy: %s", cmd.CommandPath(), VERSION, COMMIT, DATE, BUILT_BY)
		} else {
			fmt.Printf("%s: %s", cmd.CommandPath(), VERSION)
		}
		return nil
	},
}
