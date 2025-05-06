package cli2

import (
	"fmt"
	"slices"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice/status"
	"github.com/spf13/cobra"
)

func devicecmd() *cobra.Command {
	deviceCmd := &cobra.Command{
		Use:   "device",
		Short: "Command used for management of naisdevice",
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Adjust or view the naisdevice configuration",
	}
	deviceCmd.AddCommand(configCmd)

	configGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Gets the current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("er det noe her?")
		}, // naisdeviceconfigget.Action,
	}
	configCmd.AddCommand(configGetCmd)

	configSetCmd := &cobra.Command{
		Use:   "set setting value",
		Short: "Sets a configuration value",
		// Before:    naisdeviceconfigset.Before,
		// Run:       naisdeviceconfigset.Action,
	}
	configCmd.AddCommand(configSetCmd)

	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Creates a naisdevice connection, will lock until connection",
		// Run:   naisdeviceconnect.Action,
	}
	configCmd.AddCommand(connectCmd)

	disconnectCmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnects your naisdevice",
		// Run:   naisdevicedisconnect.Action,
	}
	deviceCmd.AddCommand(disconnectCmd)

	jitaCmd := &cobra.Command{
		Use:   "jita gateway-name",
		Short: "Connects to a JITA gateway",
		// Before:    naisdevicejita.Before,
		// Run:       naisdevicejita.Action,
	}
	deviceCmd.AddCommand(jitaCmd)

	statusFlags := status.Flags{}
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Shows the status of your naisdevice",

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !slices.Contains([]string{"yaml", "json"}, statusFlags.Output) {
				metrics.AddOne(cmd.Context(), "status_file_format_error_total")
				return fmt.Errorf("%v is not an implemented format", statusFlags.Output)
			}
			return nil
		},
		// Run: naisdevicestatus.Action,
	}
	statusCmd.Flags().StringVarP(&statusFlags.Output, "output", "o", "yaml", "Output format (yaml or json)")
	statusCmd.Flags().BoolVarP(&statusFlags.Verbose, "verbose", "v", false, "Verbose output")
	statusCmd.Flags().BoolVarP(&statusFlags.Quiet, "quiet", "q", false, "Quiet output")
	deviceCmd.AddCommand(statusCmd)

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Examine the health of your naisdevice",
		// Run:   naisdevicedoctor.Action,
	}
	deviceCmd.AddCommand(doctorCmd)

	return deviceCmd
}
