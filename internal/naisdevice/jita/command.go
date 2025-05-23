package jita

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type Arguments struct {
	Gateways []string
}

func Gateways(ctx context.Context) ([]string, error) {
	privilegedGateways, err := GetPrivilegedGateways(ctx)
	if err != nil {
		return nil, err
	}

	return privilegedGateways, nil
}

func Run(ctx context.Context, args Arguments) error {
	privilegedGateways, err := GetPrivilegedGateways(ctx)
	if err != nil {
		return err
	}

	for _, gateway := range args.Gateways {
		if !slices.Contains(privilegedGateways, gateway) {
			return fmt.Errorf("%v is not one of the privileged gateways: %v", gateway, strings.Join(privilegedGateways, ", "))
		}
	}

	for _, gateway := range args.Gateways {
		if err := AccessPrivilegedGateway(gateway); err != nil {
			return fmt.Errorf("access JITA gateway: %w", err)
		}
	}

	return nil
}
