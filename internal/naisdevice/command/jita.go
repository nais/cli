package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func jitacmd(parentFlags *root.Flags) *cli.Command {
	return cli.NewCommand("jita", "Connect to a JITA gateway.",
		cli.WithArgs("gateway"),
		cli.WithRun(run),
		cli.WithValidate(cli.ValidateMinArgs(1)),
		cli.WithAutoComplete(autocomplete),
	)
}

type Arguments struct {
	Gateways []string
}

func Gateways(ctx context.Context) ([]string, error) {
	privilegedGateways, err := naisdevice.GetPrivilegedGateways(ctx)
	if err != nil {
		return nil, err
	}

	return privilegedGateways, nil
}

func autocomplete(ctx context.Context, args []string, _ string) ([]string, string) {
	gateways, err := Gateways(ctx)
	if err != nil {
		msg := fmt.Sprintf("error listing gateways: %v - is it running?", err)
		return nil, msg
	}

	// don't suggest gateways already present in args
	gateways = slices.DeleteFunc(gateways, func(gateway string) bool {
		return slices.Contains(args, gateway)
	})

	return gateways, ""
}

func run(ctx context.Context, _ output.Output, args []string) error {
	privilegedGateways, err := naisdevice.GetPrivilegedGateways(ctx)
	if err != nil {
		return err
	}

	for _, gateway := range args {
		if !slices.Contains(privilegedGateways, gateway) {
			return fmt.Errorf("%v is not one of the privileged gateways: %v", gateway, strings.Join(privilegedGateways, ", "))
		}
	}

	for _, gateway := range args {
		if err := naisdevice.AccessPrivilegedGateway(gateway); err != nil {
			return fmt.Errorf("access JITA gateway: %w", err)
		}
	}

	return nil
}
