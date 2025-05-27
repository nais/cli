package naisdevice

import (
	"context"

	"github.com/nais/device/pkg/pb"
)

func GetConfig(ctx context.Context) (*pb.AgentConfiguration, error) {
	connection, err := AgentConnection()
	if err != nil {
		return nil, err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
	if err != nil {
		return nil, FormatGrpcError(err)
	}

	return configResponse.Config, nil
}
