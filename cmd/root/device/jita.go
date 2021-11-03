package device

import (
	"fmt"
	"strings"

	"github.com/nais/device/pkg/device-agent/open"
	"github.com/spf13/cobra"
)

var jitaCmd = &cobra.Command{
	Use:     "jita",
	Short:   "Connects to a JITA gateway",
	Example: `nais device jita [gateway]`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("missing required arguments: gateway")
		}
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		gateway := strings.TrimSpace(args[0])

		err := accessPrivilegedGateway(gateway)
		if err != nil {
			return err
		}
		return nil
	},
}

func accessPrivilegedGateway(gatewayName string) error {
	return open.Open(fmt.Sprintf("https://naisdevice-jita.nais.io/?gateway=%s", gatewayName))

}
