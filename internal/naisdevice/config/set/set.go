package set

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisdevice/naisdevicegrpc"
	"github.com/nais/device/pkg/pb"
)

var (
	allowedSettings = []string{"AutoConnect"}
	hiddenSettings  = []string{"ILoveNinetiesBoybands"}
)

func GetSettingValues(setting string) map[string]string {
	switch strings.ToLower(setting) {
	case "autoconnect":
		return map[string]string{
			"true":  "Enable autoconnect",
			"false": "Disable autoconnect",
		}
	case "iloveninetiesboybands":
		return map[string]string{
			"true":  "Enable tenant switching",
			"false": "Disable tenant switching",
		}
	default:
		return nil
	}
}

func GetAllowedSettings(withHidden, lowerCase bool) []string {
	settings := allowedSettings

	if withHidden {
		settings = append(settings, hiddenSettings...)
	}

	if lowerCase {
		for i, setting := range settings {
			settings[i] = strings.ToLower(setting)
		}
	}

	return settings
}

func set(ctx context.Context, setting string, value bool) error {
	connection, err := naisdevicegrpc.AgentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	// we have to fetch the agent configuration and mutate it in here. :(
	// SetAgentConfiguration on the agent's side replaces its config with the payload we send
	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
	if err != nil {
		return naisdevicegrpc.FormatGrpcError(err)
	}

	switch strings.ToLower(setting) {
	case "autoconnect":
		configResponse.Config.AutoConnect = value
	case "iloveninetiesboybands":
		configResponse.Config.ILoveNinetiesBoybands = value
	default:
		return fmt.Errorf("setting must be one of [%v]", strings.Join(GetAllowedSettings(false, false), ", "))
	}

	setConfigRequest := &pb.SetAgentConfigurationRequest{Config: configResponse.Config}
	_, err = client.SetAgentConfiguration(ctx, setConfigRequest)
	if err != nil {
		return naisdevicegrpc.FormatGrpcError(err)
	}

	return nil
}
