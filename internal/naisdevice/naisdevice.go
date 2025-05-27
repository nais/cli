package naisdevice

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nais/cli/internal/output"
	"github.com/nais/device/pkg/config"
	"github.com/nais/device/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func AgentConnection() (*grpc.ClientConn, error) {
	userConfigDir, err := config.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("naisdevice config directory: %v", err)
	}
	socket := filepath.Join(userConfigDir, "agent.sock")

	connection, err := grpc.NewClient(
		"unix:"+socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, FormatGrpcError(err)
	}

	return connection, nil
}

func FormatGrpcError(err error) error {
	gerr, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch gerr.Code() {
	case codes.Unavailable:
		return fmt.Errorf("unable to connect to naisdevice; make sure naisdevice is running")
	}
	return fmt.Errorf("%s: %s", gerr.Code(), gerr.Message())
}

func Connect(ctx context.Context, out output.Output) error {
	connection, err := AgentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer func() { _ = connection.Close() }()

	_, err = client.Login(ctx, &pb.LoginRequest{})
	if err != nil {
		return FormatGrpcError(err)
	}

	return waitForConnectionState(ctx, client, pb.AgentState_Connected, out)
}

func Disconnect(ctx context.Context, out output.Output) error {
	connection, err := AgentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer func() { _ = connection.Close() }()

	_, err = client.Logout(ctx, &pb.LogoutRequest{})
	if err != nil {
		return FormatGrpcError(err)
	}

	return waitForConnectionState(ctx, client, pb.AgentState_Disconnected, out)
}

func waitForConnectionState(ctx context.Context, client pb.DeviceAgentClient, wantedAgentState pb.AgentState, out output.Output) error {
	stream, err := client.Status(ctx, &pb.AgentStatusRequest{
		KeepConnectionOnComplete: true,
	})
	if err != nil {
		return FormatGrpcError(err)
	}

	for stream.Context().Err() == nil {
		st, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("error while receiving status: %w", err)
		}
		out.Printf("State: %s\n", st.ConnectionState)
		if st.ConnectionState == wantedAgentState {
			return nil
		}
	}

	return stream.Context().Err()
}
