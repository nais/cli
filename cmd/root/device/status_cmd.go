package device

import (
	"encoding/json"
	"fmt"

	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var statusCmd = &cobra.Command{
	Use:     "status [flags]",
	Short:   "Shows the status of your naisdevice",
	Example: `nais device status -q | nais device status -o yaml`,
	RunE: func(command *cobra.Command, args []string) error {
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		connection, err := agentConnection()
		if err != nil {
			return formatGrpcError(err)
		}

		client := pb.NewDeviceAgentClient(connection)
		defer connection.Close()

		stream, err := client.Status(command.Context(), &pb.AgentStatusRequest{
			KeepConnectionOnComplete: true,
		})
		if err != nil {
			return formatGrpcError(err)
		}

		status, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("receive status: %w", err)
		}

		if viper.GetString(OutputFlag) == "yaml" {
			out, err := yaml.Marshal(status)
			if err != nil {
				return fmt.Errorf("Marshaling status: %v", err)
			}
			fmt.Println(string(out))
			return nil
		} else if viper.GetString(OutputFlag) == "json" {
			out, err := json.Marshal(status)
			if err != nil {
				return fmt.Errorf("Marshaling status: %v", err)
			}
			fmt.Println(string(out))
			return nil
		}

		if viper.GetBool(QuietFlag) {
			fmt.Println(status.ConnectionState.String())
			return nil
		}

		fmt.Printf("naisdevice status: %s\n", status.ConnectionStateString())
		if status.NewVersionAvailable {
			fmt.Printf("\nNew version of naisdevice available!\nSee https://doc.nais.io/device/update for upgrade instructions.\n")
		}

		healthy := func(gw *pb.Gateway) string {
			if gw.Healthy {
				return "connected"
			} else {
				return "disconnected"
			}
		}

		privileged := func(gw *pb.Gateway) string {
			if gw.RequiresPrivilegedAccess {
				if gw.Healthy {
					return "active"
				}
				return "required"
			} else {
				return ""
			}
		}

		if len(status.Gateways) > 0 {
			fmt.Printf("\n%-30s\t%-15s\t%-15s\n", "GATEWAY", "STATE", "JITA")
		}
		for _, gw := range status.Gateways {
			fmt.Printf("%-30s\t%-15s\t%-15s\n", gw.Name, healthy(gw), privileged(gw))
		}

		return nil
	},
}
