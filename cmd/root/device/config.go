package device

import "github.com/spf13/cobra"

type DeviceConfig struct {
	device     *cobra.Command
	connect    *cobra.Command
	disconnect *cobra.Command
	status     *cobra.Command
}

func NewDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		device:     deviceCmd,
		connect:    connectCmd,
		disconnect: disconnectCmd,
		status:     statusCmd,
	}
}

func (d DeviceConfig) InitCmds(root *cobra.Command) {
	root.AddCommand(d.device)
	d.device.AddCommand(d.connect)
	d.device.AddCommand(d.disconnect)
	d.device.AddCommand(d.status)
}
