package device

import (
	"fmt"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   "Creates a naisdevice connection",
	Example: `nais device connect`,
	RunE: func(command *cobra.Command, args []string) error {
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		connection, err := agentConnection()
		if err != nil {
			return fmt.Errorf("Agent connection: %v", err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		_, err = client.Login(command.Context(), &pb.LoginRequest{})
		if err != nil {
			return fmt.Errorf("Connecting to naisdevice. Ensure that naisdevice is running.\n: %v", err)
		}

		stream, err := client.Status(command.Context(), &pb.AgentStatusRequest{
			KeepConnectionOnComplete: true,
		})
		if err != nil {
			return fmt.Errorf("Connecting to naisdevice. Ensure that naisdevice is running.\n%v", err)
		}

		for stream.Context().Err() == nil {
			status, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("receive status: %w", err)
			}
			fmt.Printf("state: %s\n", status.ConnectionState)
			if status.ConnectionState == pb.AgentState_Connected {
				return nil
			}
		}

		return stream.Context().Err()
	},
}
