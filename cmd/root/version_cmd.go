package root

import (
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version [flags]",
	Short: "Show 'nais-cli' client version",
	RunE: func(command *cobra.Command, args []string) error {
		result, err := command.Flags().GetBool(cmd.CommitInformation)
		if err != nil {
			return err
		}

		if result {
			fmt.Printf("%s: %s commit: %s date: %s builtBy: %s\n", command.CommandPath(), VERSION, COMMIT, DATE, BuiltBy)
		} else {
			fmt.Printf("%s: %s\n", command.CommandPath(), VERSION)
		}
		return nil
	},
}
