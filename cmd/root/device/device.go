package device

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/nais/device/pkg/config"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	connectedStatus    = "Connected"
	disconnectedStatus = "Disconnected"
)

var deviceCmd = &cobra.Command{
	Use:   "device [command] [args] [flags]",
	Short: "Command used for management of 'naisdevice'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required commands")
	},
}

func agentConnection() (*grpc.ClientConn, error) {
	userConfigDir, err := config.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("naisdevice config directory: %v", err)
	}
	socket := filepath.Join(userConfigDir, "agent.sock")

	return grpc.Dial(
		"unix:"+socket,
		grpc.WithInsecure(),
	)
}

func waitForStatus(desired string, timeout time.Duration) error {
	stopTrying := time.Now().Add(timeout)
	for {
		state, err := status()
		if state == desired {
			break
		}
		if err != nil {
			return fmt.Errorf("Getting status: %v", err)
		}
		if time.Now().After(stopTrying) {
			return fmt.Errorf("Timed out")
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
