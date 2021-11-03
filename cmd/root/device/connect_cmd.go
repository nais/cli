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
		state, err := status()
		if err != nil {
			return fmt.Errorf("Getting status: %v", err)
		}
		if state != disconnectedStatus {
			return nil
		}
		connection, err := agentConnection()
		if err != nil {
			return fmt.Errorf("Agent connection: %v", err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		_, err = client.Login(command.Context(), &pb.LoginRequest{})
		if err != nil {
			return fmt.Errorf("Connecting to naisdevice: %v", err)
		}
		return nil
	},
}
