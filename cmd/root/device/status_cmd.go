package device

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/nais/device/pkg/config"
	"github.com/spf13/cobra"
)

func status() (string, error) {
	userConfigDir, err := config.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("naisdevice config directory: %v", err)
	}
	statusFile := filepath.Join(userConfigDir, "agent_status")
	file, err := ioutil.ReadFile(statusFile)
	if err != nil {
		return "", fmt.Errorf("status file: %v", err)
	}
	return string(file), nil
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Shows the status of your naisdevice",
	Example: `nais device status`,
	RunE: func(command *cobra.Command, args []string) error {
		state, err := status()
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", state)
		return nil
	},
}
