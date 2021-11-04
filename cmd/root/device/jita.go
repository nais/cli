package device

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/device/pkg/device-agent/open"
	"github.com/nais/device/pkg/pb"
	"github.com/spf13/cobra"
)

var jitaCmd = &cobra.Command{
	Use:     "jita [gateway]",
	Short:   "Connects to a JITA gateway",
	Example: `nais device jita postgres-prod`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getGatewayList(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("missing required arguments: gateway")
		}
		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		gateway := strings.TrimSpace(args[0])

		err := accessPrivilegedGateway(gateway)
		if err != nil {
			return err
		}
		return nil
	},
}

func accessPrivilegedGateway(gatewayName string) error {
	return open.Open(fmt.Sprintf("https://naisdevice-jita.nais.io/?gateway=%s", gatewayName))

}

func getGatewayList(ctx context.Context, toComplete string) []string {
	connection, err := agentConnection()
	if err != nil {
		return nil
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	stream, err := client.Status(ctx, &pb.AgentStatusRequest{
		KeepConnectionOnComplete: true,
	})

	status, err := stream.Recv()
	if err != nil {
		return nil
	}

	ret := make([]string, 0)
	for _, gw := range status.Gateways {
		if !gw.Healthy && gw.RequiresPrivilegedAccess && strings.Contains(gw.Name, strings.ToLower(toComplete)) {
			ret = append(ret, gw.Name)
		}
	}
	return ret
}
