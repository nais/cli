package get

import (
	"context"

	"github.com/nais/cli/internal/naisdevice/naisdevicegrpc"
	"github.com/nais/device/pkg/pb"
)

func get(ctx context.Context) (*pb.AgentConfiguration, error) {
	connection, err := naisdevicegrpc.AgentConnection()
	if err != nil {
		return nil, err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
	if err != nil {
		return nil, naisdevicegrpc.FormatGrpcError(err)
	}

	return configResponse.Config, nil
}
