package device

import (
	"fmt"
	"time"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:     "disconnect",
	Short:   "Disconnects your naisdevice",
	Example: `nais device disconnect`,
	RunE: func(command *cobra.Command, args []string) error {
		state, err := status()
		if err != nil {
			return fmt.Errorf("Getting status: %v", err)
		}
		if state != connectedStatus {
			return nil
		}
		connection, err := agentConnection()
		if err != nil {
			return fmt.Errorf("Agent connection: %v", err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		_, err = client.Logout(command.Context(), &pb.LogoutRequest{})
		if err != nil {
			return fmt.Errorf("Disconnecting from naisdevice. Ensure that naisdevice is running.\n%v", err)
		}
		err = waitForStatus(disconnectedStatus, 30*time.Second)
		if err != nil {
			return fmt.Errorf("Waiting for disconnected state: %v", err)
		}
		fmt.Println(disconnectedStatus)
		return nil
	},
}
