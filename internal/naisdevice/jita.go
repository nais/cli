package naisdevice

import (
	"context"

	"github.com/nais/device/pkg/pb"
)

func ShowJITA(ctx context.Context, gateway string) error {
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
