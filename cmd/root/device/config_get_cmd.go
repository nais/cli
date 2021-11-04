package device

import (
	"fmt"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var configGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Gets the current configuration",
	Example: `nais device config get`,
	RunE: func(command *cobra.Command, args []string) error {
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		connection, err := agentConnection()
		if err != nil {
			return formatGrpcError(err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		configResponse, err := client.GetAgentConfiguration(command.Context(), &pb.GetAgentConfigurationRequest{})
		if err != nil {
			return formatGrpcError(err)
		}

		fmt.Printf("AutoConnect:\t%v\n", configResponse.Config.AutoConnect)
		fmt.Printf("CertRenewal:\t%v\n", configResponse.Config.CertRenewal)

		return nil
	},
}
