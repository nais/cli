package naisdevice

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/nais/device/pkg/pb"
	"gopkg.in/yaml.v3"
)

func GetStatus(ctx context.Context) (*pb.AgentStatus, error) {
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
		return nil, FormatGrpcError(err)
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
	fmt.Printf("Naisdevice status: %s\n", status.ConnectionStateString())
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

func IsConnected(ctx context.Context) bool {
	agentStatus, err := GetStatus(ctx)
	if err != nil {
		return false
	}
	return agentStatus.GetConnectionState() == pb.AgentState_Connected
}
