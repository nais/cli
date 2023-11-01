package naisdevice

import (
	"context"
	"fmt"

	"github.com/nais/device/pkg/device-agent/open"
	"github.com/nais/device/pkg/pb"
)

func AccessPrivilegedGateway(gatewayName string) {
	open.Open(fmt.Sprintf("https://naisdevice-jita.external.prod-gcp.nav.cloud.nais.io/?gateway=%s", gatewayName))
}

func GetPrivilegedGateways(ctx context.Context) ([]string, error) {
	connection, err := agentConnection()
	if err != nil {
		return nil, err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

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

	gateways := make([]string, 0)
	for _, gateway := range status.Gateways {
		if !gateway.Healthy && gateway.RequiresPrivilegedAccess {
			gateways = append(gateways, gateway.Name)
		}
	}
	return gateways, nil
}
