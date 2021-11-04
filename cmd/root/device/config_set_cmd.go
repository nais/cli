package device

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:               "set [setting] [value]",
	Short:             "Sets a configuration value",
	Example:           `nais device config set AutoConnect true`,
	ValidArgsFunction: validSettings,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("missing required arguments: setting, value")
		}
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		setting := strings.TrimSpace(args[0])
		valueString := strings.TrimSpace(args[1])

		value, err := strconv.ParseBool(valueString)
		if err != nil {
			return fmt.Errorf("Parsing setting as boolean value: %v", err)
		}

		connection, err := agentConnection()
		if err != nil {
			return formatGrpcError(err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		// we have to fetch the agent configuration and mutate it in here. :(
		// SetAgentConfiguration on the agent's side replaces its config with the payload we send
		configResponse, err := client.GetAgentConfiguration(command.Context(), &pb.GetAgentConfigurationRequest{})
		if err != nil {
			return formatGrpcError(err)
		}

		switch strings.ToLower(setting) {
		case "autoconnect":
			configResponse.Config.AutoConnect = value
		case "certrenewal":
			configResponse.Config.CertRenewal = value
		default:
			return fmt.Errorf("Setting must be one of [autoconnect, certrenewal]")
		}

		setConfigRequest := &pb.SetAgentConfigurationRequest{Config: configResponse.Config}
		_, err = client.SetAgentConfiguration(command.Context(), setConfigRequest)
		if err != nil {
			return formatGrpcError(err)
		}

		return nil
	},
}

func validSettings(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ret := make([]string, 0)

	// one could make something based on `pb.AgentConfiguration{}`
	// and reflect and find the public parameters and put that into `ret`.. but... maybe not
	allowedList := []string{"AutoConnect", "CertRenewal"}
	for _, allowed := range allowedList {
		if strings.Contains(strings.ToLower(allowed), strings.ToLower(toComplete)) {
			ret = append(ret, allowed)
		}
	}

	return ret, cobra.ShellCompDirectiveNoFileComp
}
