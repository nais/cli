package device

import (
	"fmt"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:     "disconnect",
	Short:   "Disconnects your naisdevice",
	Example: `nais device disconnect`,
	RunE: func(command *cobra.Command, args []string) error {
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		connection, err := agentConnection()
		if err != nil {
			return formatGrpcError(err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		_, err = client.Logout(command.Context(), &pb.LogoutRequest{})
		if err != nil {
			return formatGrpcError(err)
		}

		stream, err := client.Status(command.Context(), &pb.AgentStatusRequest{
			KeepConnectionOnComplete: true,
		})

		if err != nil {
			return formatGrpcError(err)
		}

		for stream.Context().Err() == nil {
			status, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("receive status: %w", err)
			}
			fmt.Printf("state: %s\n", status.ConnectionState)
			if status.ConnectionState == pb.AgentState_Disconnected {
				return nil
			}
		}

		return stream.Context().Err()
	},
}
