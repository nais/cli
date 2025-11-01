package naisdevice

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/device/pkg/pb"
	"github.com/nais/naistrix"
)

func GetStatus(ctx context.Context) (*pb.AgentStatus, error) {
	connection, err := AgentConnection()
	if err != nil {
		return nil, err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer func() { _ = connection.Close() }()

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
	}

	return "disconnected"
}

func gatewayPrivileged(gw *pb.Gateway) string {
	if gw.RequiresPrivilegedAccess {
		if gw.Healthy {
			return "active"
		}
		return "required"
	}
	return ""
}

func PrintVerboseStatus(status *pb.AgentStatus, out *naistrix.OutputWriter) {
	out.Printf("Naisdevice status: %s\n", status.ConnectionStateString())
	if status.NewVersionAvailable {
		out.Printf("\nNew version of naisdevice available!\nSee https://doc.nais.io/device/update for upgrade instructions.\n")
	}

	if len(status.Gateways) > 0 {
		out.Printf("\n%-30s\t%-15s\t%-15s\n", "GATEWAY", "STATE", "JITA")
	}

	sort.Slice(status.Gateways, func(i, j int) bool {
		return status.Gateways[i].Name < status.Gateways[j].Name
	})

	for _, gw := range status.Gateways {
		out.Printf("%-30s\t%-15s\t%-15s\n", gw.Name, gatewayHealthy(gw), gatewayPrivileged(gw))
	}
}

func PrintFormattedStatus(format string, status *pb.AgentStatus, out *naistrix.OutputWriter) error {
	var o interface {
		Render(v any) error
	}

	switch format {
	case "yaml":
		o = out.YAML()
	case "json":
		o = out.JSON()
	default:
		return fmt.Errorf("unknown format: %q", format)
	}

	return o.Render(status)
}
