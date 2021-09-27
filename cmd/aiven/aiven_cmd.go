package aiven

import (
	"fmt"
	"github.com/spf13/cobra"
)

var AivenCommand = &cobra.Command{
	Use:   "aiven [command] [args] [flags]",
	Short: "Command used for management of 'AivenApplication'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required commands")
	},
}
