package naisdevice

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nais/device/pkg/config"
	"github.com/nais/device/pkg/pb"
	"github.com/nais/naistrix"
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

func Connect(ctx context.Context, wait bool, out *naistrix.OutputWriter) error {
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

	if err := waitForConnectionState(ctx, client, pb.AgentState_Connected, out); err != nil {
		return err
	}

	if !wait {
		return nil
	}

	return waitForHealthyGateways(ctx, client, out)
}

func Disconnect(ctx context.Context, out *naistrix.OutputWriter) error {
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

func waitForConnectionState(ctx context.Context, client pb.DeviceAgentClient, wantedAgentState pb.AgentState, out *naistrix.OutputWriter) error {
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

// waitForHealthyGateways blocks until all gateways not requiring privileged access (JITA) are healthy.
func waitForHealthyGateways(ctx context.Context, client pb.DeviceAgentClient, out *naistrix.OutputWriter) error {
	stream, err := client.Status(ctx, &pb.AgentStatusRequest{
		KeepConnectionOnComplete: true,
	})
	if err != nil {
		return FormatGrpcError(err)
	}

	out.Println("Waiting for gateways to become healthy...")
	lastPending := -1
	for stream.Context().Err() == nil {
		st, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("error while receiving status: %w", err)
		}

		if st.ConnectionState != pb.AgentState_Connected {
			continue
		}

		total, pending := 0, 0
		for _, gw := range st.Gateways {
			if gw.RequiresPrivilegedAccess {
				continue
			}
			total++
			if !gw.Healthy {
				pending++
			}
		}

		// Gateway list may be empty right after connecting; wait until populated.
		if total == 0 {
			continue
		}

		if pending == 0 {
			out.Printf("All %d gateways healthy\n", total)
			return nil
		}

		if pending != lastPending {
			out.Printf("Gateways healthy: %d/%d\n", total-pending, total)
			lastPending = pending
		}
	}

	return stream.Context().Err()
}
