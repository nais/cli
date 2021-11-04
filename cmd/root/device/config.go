package device

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	QuietFlag       = "quiet"
	QuietFlagShort  = "q"
	OutputFlag      = "output"
	OutputFlagShort = "o"
)

type DeviceConfig struct {
	device     *cobra.Command
	connect    *cobra.Command
	disconnect *cobra.Command
	status     *cobra.Command
	jita       *cobra.Command
	config     *cobra.Command
	configGet  *cobra.Command
	configSet  *cobra.Command
}

func NewDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		device:     deviceCmd,
		connect:    connectCmd,
		disconnect: disconnectCmd,
		status:     statusCmd,
		jita:       jitaCmd,
		config:     configCmd,
		configGet:  configGetCmd,
		configSet:  configSetCmd,
	}
}

func (d DeviceConfig) InitCmds(root *cobra.Command) {
	root.AddCommand(d.device)
	d.device.AddCommand(d.connect)
	d.device.AddCommand(d.disconnect)
	d.device.AddCommand(d.status)
	d.device.AddCommand(d.jita)
	d.device.AddCommand(d.config)
	d.config.AddCommand(d.configGet)
	d.config.AddCommand(d.configSet)
	d.status.Flags().BoolP(QuietFlag, QuietFlagShort, false, "Reduce verbosity.")
	viper.BindPFlag(QuietFlag, d.status.Flag(QuietFlag))
	d.status.Flags().StringP(OutputFlag, OutputFlagShort, "", "Output format")
	viper.BindPFlag(OutputFlag, d.status.Flag(OutputFlag))
}
