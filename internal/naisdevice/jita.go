package naisdevice

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/urlopen"
	"github.com/nais/device/pkg/pb"
)

func AccessPrivilegedGateway(gatewayName string) error {
	url := fmt.Sprintf("https://naisdevice-jita.external.prod-gcp.nav.cloud.nais.io/?gateway=%s", gatewayName)
	err := urlopen.Open(url)
	if err != nil {
		return fmt.Errorf("unable to open your browser, please open this manually: %s", url)
	}
	return nil
}

func GetPrivilegedGateways(ctx context.Context) ([]string, error) {
	connection, err := AgentConnection()
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
