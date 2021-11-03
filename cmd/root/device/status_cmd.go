package device

import (
	"fmt"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Shows the status of your naisdevice",
	Example: `nais device status`,
	RunE: func(command *cobra.Command, args []string) error {
		connection, err := agentConnection()
		if err != nil {
			return fmt.Errorf("Agent connection: %v", err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		stream, err := client.Status(command.Context(), &pb.AgentStatusRequest{
			KeepConnectionOnComplete: true,
		})
		if err != nil {
			return fmt.Errorf("Connecting to naisdevice. Ensure that naisdevice is running.\n%v", err)
		}

		status, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("receive status: %w", err)
		}

		fmt.Println(status.ConnectionState.String())

		return nil
	},
}
