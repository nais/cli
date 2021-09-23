package aiven

import (
	"fmt"
	"github.com/spf13/cobra"
)

var AivenCommand = &cobra.Command{
	Use:   "aiven [command] [args] [flags]",
	Short: "Create a protected & time-limited aivenApplication",
	Long:  `This command will apply a aivenApplication based on information given and aivenator creates a set of credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required commands")
	},
}
