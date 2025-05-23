package get

// import (
// 	"context"
//
// 	"github.com/nais/cli/internal/naisdevice"
// 	"github.com/nais/device/pkg/pb"
// )
//
// func get(ctx context.Context) (*pb.AgentConfiguration, error) {
// 	connection, err := naisdevice.AgentConnection()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	client := pb.NewDeviceAgentClient(connection)
// 	defer connection.Close()
//
// 	configResponse, err := client.GetAgentConfiguration(ctx, &pb.GetAgentConfigurationRequest{})
// 	if err != nil {
// 		return nil, naisdevice.FormatGrpcError(err)
// 	}
//
// 	return configResponse.Config, nil
// }
