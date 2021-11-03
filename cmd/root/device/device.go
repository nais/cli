package device

import (
	"fmt"
	"path/filepath"

	"github.com/nais/device/pkg/config"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var deviceCmd = &cobra.Command{
	Use:   "device [command] [args] [flags]",
	Short: "Command used for management of 'naisdevice'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required command")
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
