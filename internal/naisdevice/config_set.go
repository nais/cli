package naisdevice

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/device/pkg/pb"
	"github.com/nais/naistrix"
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

func SetConfig(ctx context.Context, setting string, value bool) error {
	connection, err := AgentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	// we have to fetch the agent configuration and mutate it in here. :(
	// SetAgentConfiguration on the agent's side replaces its config with the payload we send
	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
	if err != nil {
		return FormatGrpcError(err)
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
		return FormatGrpcError(err)
	}

	return nil
}

func AutocompleteSet(_ context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
	if args.Len() == 0 {
		return GetAllowedSettings(false, false), ""
	} else if args.Len() == 1 {
		var completions []string
		for key, value := range GetSettingValues(args.Get("setting")) {
			completions = append(completions, key+"\t"+value)
		}
		return completions, "Possible values"
	}

	return nil, "no more inputs expected, press enter"
}
