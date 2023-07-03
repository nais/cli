package naisdevice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nais/device/pkg/pb"
	"gopkg.in/yaml.v3"
	"sort"
)

func Connect(ctx context.Context) error {
	connection, err := agentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	_, err = client.Login(ctx, &pb.LoginRequest{})
	if err != nil {
		return formatGrpcError(err)
	}

	return waitForConnectionState(ctx, client, pb.AgentState_Connected)
}

func Disconnect(ctx context.Context) error {
	connection, err := agentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	_, err = client.Logout(ctx, &pb.LogoutRequest{})
	if err != nil {
		return formatGrpcError(err)
	}

	return waitForConnectionState(ctx, client, pb.AgentState_Disconnected)
}

func waitForConnectionState(ctx context.Context, client pb.DeviceAgentClient, wantedAgentState pb.AgentState) error {
	stream, err := client.Status(ctx, &pb.AgentStatusRequest{
		KeepConnectionOnComplete: true,
	})

	if err != nil {
		return formatGrpcError(err)
	}

	for stream.Context().Err() == nil {
		status, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("error while receiving status: %w", err)
		}
		fmt.Printf("state: %s\n", status.ConnectionState)
		if status.ConnectionState == wantedAgentState {
			return nil
		}
	}

	return stream.Context().Err()
}

func GetStatus(ctx context.Context) (*pb.AgentStatus, error) {
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
		return nil, formatGrpcError(err)
	}

	return stream.Recv()
}

func gatewayHealthy(gw *pb.Gateway) string {
	if gw.Healthy {
		return "connected"
	} else {
		return "disconnected"
	}
}

func gatewayPrivileged(gw *pb.Gateway) string {
	if gw.RequiresPrivilegedAccess {
		if gw.Healthy {
			return "active"
		}
		return "required"
	} else {
		return ""
	}
}

func PrintVerboseStatus(status *pb.AgentStatus) {
	fmt.Printf("naisdevice status: %s\n", status.ConnectionStateString())
	if status.NewVersionAvailable {
		fmt.Printf("\nNew version of naisdevice available!\nSee https://doc.nais.io/device/update for upgrade instructions.\n")
	}

	if len(status.Gateways) > 0 {
		fmt.Printf("\n%-30s\t%-15s\t%-15s\n", "GATEWAY", "STATE", "JITA")
	}

	sort.Slice(status.Gateways, func(i, j int) bool {
		return status.Gateways[i].Name < status.Gateways[j].Name
	})

	for _, gw := range status.Gateways {
		fmt.Printf("%-30s\t%-15s\t%-15s\n", gw.Name, gatewayHealthy(gw), gatewayPrivileged(gw))
	}
}

func PrintFormattedStatus(format string, status *pb.AgentStatus) error {
	switch format {
	case "yaml":
		out, err := yaml.Marshal(status)
		if err != nil {
			return fmt.Errorf("marshaling status: %v", err)
		}
		fmt.Println(string(out))
	case "json":
		out, err := json.Marshal(status)
		if err != nil {
			return fmt.Errorf("marshaling status: %v", err)
		}
		fmt.Println(string(out))
	}

	return nil
}
