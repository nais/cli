package device

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config [command]",
	Short:   "Adjust or view the naisdevice configuration",
	Example: `nais device config set autoconnect true`,
	RunE: func(command *cobra.Command, args []string) error {
		return fmt.Errorf("missing required command")
	},
}
