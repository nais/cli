package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show debuk client version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Use + " " + VERSION)
	},
}
