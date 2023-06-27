package naisdevice

import (
	"context"
	"fmt"
	"github.com/nais/device/pkg/pb"
	"strings"
)

var (
	AllowedSettings = []string{"AutoConnect", "CertRenewal"}
)

func GetConfiguration(ctx context.Context) (*pb.AgentConfiguration, error) {
	connection, err := agentConnection()
	if err != nil {
		return nil, err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
	if err != nil {
		return nil, formatGrpcError(err)
	}

	return configResponse.Config, nil
}

func SetConfiguration(ctx context.Context, setting string, value bool) error {
	connection, err := agentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	// we have to fetch the agent configuration and mutate it in here. :(
	// SetAgentConfiguration on the agent's side replaces its config with the payload we send
	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
	if err != nil {
		return formatGrpcError(err)
	}

	switch strings.ToLower(setting) {
	case "autoconnect":
		configResponse.Config.AutoConnect = value
	case "certrenewal":
		configResponse.Config.CertRenewal = value
	default:
		return fmt.Errorf("setting must be one of [autoconnect, certrenewal]")
	}

	setConfigRequest := &pb.SetAgentConfigurationRequest{Config: configResponse.Config}
	_, err = client.SetAgentConfiguration(ctx, setConfigRequest)
	if err != nil {
		return formatGrpcError(err)
	}

	return nil
}
