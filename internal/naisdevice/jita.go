package naisdevice

import (
	"context"
	"fmt"

	"github.com/nais/device/pkg/pb"
)

func AccessPrivilegedGateway(ctx context.Context, gateway string) error {
	connection, err := AgentConnection()
	if err != nil {
		return err
	}
	defer func() { _ = connection.Close() }()

	_, err = pb.NewDeviceAgentClient(connection).ShowJita(ctx, &pb.ShowJitaRequest{
		Gateway: gateway,
	})

	return err
}

func GetPrivilegedGateways(ctx context.Context) ([]string, error) {
	connection, err := AgentConnection()
	if err != nil {
		return nil, err
	}
	defer func() { _ = connection.Close() }()

	client := pb.NewDeviceAgentClient(connection)

	stream, err := client.Status(ctx, &pb.AgentStatusRequest{
		KeepConnectionOnComplete: true,
	})
	if err != nil {
		return nil, err
	}

	status, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	if status.ConnectionState != pb.AgentState_Connected {
		return nil, fmt.Errorf("agent not connected")
	}

	gateways := make([]string, 0)
	for _, gateway := range status.Gateways {
		if !gateway.Healthy && gateway.RequiresPrivilegedAccess {
			gateways = append(gateways, gateway.Name)
		}
	}
	return gateways, nil
}
