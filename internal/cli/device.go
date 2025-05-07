package cli

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/nais/cli/internal/naisdevice/config/get"
	"github.com/nais/cli/internal/naisdevice/config/set"
	"github.com/nais/cli/internal/naisdevice/connect"
	"github.com/nais/cli/internal/naisdevice/disconnect"
	"github.com/nais/cli/internal/naisdevice/doctor"
	"github.com/nais/cli/internal/naisdevice/jita"
	"github.com/nais/cli/internal/naisdevice/status"
	"github.com/spf13/cobra"
)

func device() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Command used for management of naisdevice",
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Adjust or view the naisdevice configuration",
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Gets the current configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return get.Run(cmd.Context())
		},
	}

	setCmd := &cobra.Command{
		Use:   "set setting value",
		Short: "Sets a configuration value",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return set.GetAllowedSettings(false, false), cobra.ShellCompDirectiveDefault
			} else if len(args) == 1 {
				var completions []cobra.Completion
				for key, value := range set.GetSettingValues(args[0]) {
					completions = append(completions, cobra.CompletionWithDesc(key, value))
				}
				return cobra.AppendActiveHelp(completions, "Possible values"), cobra.ShellCompDirectiveDefault
			}
			return cobra.AppendActiveHelp(nil, "no more inputs expected, press enter"), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			setting := args[0]
			value, err := strconv.ParseBool(args[1])
			if err != nil {
				return err
			}

			arguments := set.Arguments{
				Setting: setting,
				Value:   value,
			}

			return set.Run(cmd.Context(), arguments)
		},
	}

	configCmd.AddCommand(getCmd, setCmd)

	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Creates a naisdevice connection, will lock until connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			return connect.Run(cmd.Context())
		},
	}

	statusFlags := status.Flags{}
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Shows the status of your naisdevice",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !slices.Contains([]string{"", "yaml", "json"}, statusFlags.Output) {
				// metrics.AddOne(cmd.Context(), "status_file_format_error_total")
				return fmt.Errorf("%v is not an implemented format", statusFlags.Output)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.Run(cmd.Context(), statusFlags)
		},
	}
	statusCmd.Flags().StringVarP(&statusFlags.Output, "output", "o", "", "Output format (yaml or json)")
	statusCmd.Flags().BoolVarP(&statusFlags.Verbose, "verbose", "v", false, "Verbose output")
	statusCmd.Flags().BoolVarP(&statusFlags.Quiet, "quiet", "q", false, "Quiet output")

	cmd.AddCommand(
		configCmd,
		connectCmd,
		statusCmd,
		&cobra.Command{
			Use:   "disconnect",
			Short: "Disconnects your naisdevice",
			RunE: func(cmd *cobra.Command, args []string) error {
				return disconnect.Run(cmd.Context())
			},
		},
		&cobra.Command{
			Use:   "jita gateway-name",
			Short: "Connects to a JITA gateway",
			Args:  cobra.MinimumNArgs(1),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
				gateways, err := jita.Gateways(cmd.Context())

				// don't suggest gateways already present in args
				gateways = slices.DeleteFunc(gateways, func(gateway string) bool {
					return slices.Contains(args, gateway)
				})

				if err != nil {
					return cobra.AppendActiveHelp(nil, "not connected to naisdevice - is it running?"), cobra.ShellCompDirectiveNoFileComp
				}
				return gateways, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return jita.Run(cmd.Context(), jita.Arguments{Gateways: args})
			},
		},
		&cobra.Command{
			Use:   "doctor",
			Short: "Examine the health of your naisdevice",
			RunE: func(cmd *cobra.Command, args []string) error {
				return doctor.Run(cmd.Context())
			},
		},
	)

	return cmd
}
