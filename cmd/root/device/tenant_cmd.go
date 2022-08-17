package device

import (
	"fmt"
	"strings"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var tenantCmd = &cobra.Command{
	Use:     "tenant",
	Short:   "Sets tenant for naisdevice",
	Example: `nais device tenant NAV`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("missing required arguments: tenant name")
		}
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		tenant := strings.TrimSpace(args[0])

		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		connection, err := agentConnection()
		if err != nil {
			return formatGrpcError(err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		_, err = client.SetActiveTenant(command.Context(), &pb.SetActiveTenantRequest{Name: tenant})
		if err != nil {
			return formatGrpcError(err)
		}

		fmt.Println("Set tenant to", tenant)

		return nil
	},
}
