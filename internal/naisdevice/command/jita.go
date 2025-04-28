package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func jita() *cli.Command {
	return &cli.Command{
		Name:      "jita",
		Usage:     "Connects to a JITA gateway",
		ArgsUsage: "gateway",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Args().Len() < 1 {
				metrics.AddOne(ctx, "jita_arguments_error_total")
				return ctx, fmt.Errorf("missing required arguments: gateway")
			}

			gateway := cmd.Args().First()
			privilegedGateways, err := naisdevice.GetPrivilegedGateways(ctx)
			if err != nil {
				return ctx, err
			}

			if !slices.Contains(privilegedGateways, gateway) {
				metrics.AddOne(ctx, "device_gateway_error_total")
				return ctx, fmt.Errorf("%v is not one of the privileged gateways: %v", gateway, strings.Join(privilegedGateways, ", "))
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			gateway := cmd.Args().First()
			return naisdevice.AccessPrivilegedGateway(gateway)
		},
	}
}
