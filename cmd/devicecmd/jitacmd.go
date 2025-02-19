package devicecmd

import (
	"fmt"
	"strings"

	"github.com/nais/cli/pkg/metrics"

	"github.com/urfave/cli/v2"
	"k8s.io/utils/strings/slices"

	"github.com/nais/cli/pkg/naisdevice"
)

func jitaCommand() *cli.Command {
	return &cli.Command{
		Name:      "jita",
		Usage:     "Connects to a JITA gateway",
		ArgsUsage: "gateway",
		Before: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				metrics.AddOne("jita_arguments_error_total")
				return fmt.Errorf("missing required arguments: gateway")
			}

			gateway := context.Args().First()
			privilegedGateways, err := naisdevice.GetPrivilegedGateways(context.Context)
			if err != nil {
				return err
			}

			if !slices.Contains(privilegedGateways, gateway) {
				metrics.AddOne("device_gateway_error_total")
				return fmt.Errorf("%v is not one of the privileged gateways: %v", gateway, strings.Join(privilegedGateways, ", "))
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			gateway := context.Args().First()
			return naisdevice.AccessPrivilegedGateway(gateway)
		},
	}
}
