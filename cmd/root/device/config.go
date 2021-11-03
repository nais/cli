package device

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	QuietFlag      = "quiet"
	QuietFlagShort = "q"
)

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
	d.status.Flags().BoolP(QuietFlag, QuietFlagShort, false, "Reduce verbosity.")
	viper.BindPFlag(QuietFlag, d.status.Flag(QuietFlag))
}
