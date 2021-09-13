package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show 'nais-d' client version",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := cmd.Flags().GetBool(CommitInformation)
		if err != nil {
			return err
		}

		if result {
			fmt.Printf("%s: %s commit: %s date: %s builtBy: %s", cmd.Use, VERSION, COMMIT, DATE, BUILT_BY)
		} else {
			fmt.Printf("%s: %s", cmd.Use, VERSION)
		}
		return nil
	},
}
