package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/device/pkg/pb"
	"github.com/nais/naistrix"
)

type YesNo bool

func (y YesNo) String() string {
	if y {
		return "Yes"
	}
	return "No"
}

type gw struct {
	Name      string `json:"name"`
	Connected YesNo  `json:"connected"`
	Jita      YesNo  `heading:"Requires JITA" json:"jita"`
}

func gatewaycmd(parentFlags *flag.Device) *naistrix.Command {
	flags := &flag.Gateway{Device: parentFlags}

	return &naistrix.Command{
		Name:        "gateway",
		Title:       "Interact with naisdevice gateways.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listcommand(flags),
			describecommand(flags),
			connectcommand(flags),
		},
	}
}

func listcommand(parentFlags *flag.Gateway) *naistrix.Command {
	flags := &flag.List{Gateway: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List gateways.",
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			agentStatus, err := naisdevice.GetStatus(ctx)
			if err != nil {
				return err
			}

			if agentStatus.GetConnectionState() != pb.AgentState_Connected {
				return fmt.Errorf("not connected to naisdevice")
			}

			allGateways := agentStatus.GetGateways()
			gateways := make([]gw, len(allGateways))
			for i, g := range allGateways {
				gateways[i] = gw{
					Name:      g.Name,
					Connected: YesNo(g.Healthy),
					Jita:      YesNo(g.RequiresPrivilegedAccess),
				}
			}

			var o interface {
				Render(v any) error
			}

			switch flags.Output {
			case "yaml":
				o = out.YAML()
			case "json":
				o = out.JSON()
			default:
				o = out.Table()
			}

			return o.Render(gateways)
		},
	}
}

func describecommand(parentFlags *flag.Gateway) *naistrix.Command {
	flags := &flag.Describe{Gateway: parentFlags}
	return &naistrix.Command{
		Name:  "describe",
		Title: "Describe a gateway.",
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "gateway"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return nil
		},
	}
}

func connectcommand(parentFlags *flag.Gateway) *naistrix.Command {
	flags := &flag.Connect{Gateway: parentFlags}
	return &naistrix.Command{
		Name:  "connect",
		Title: "Connect to a gateway.",
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "gateway"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return nil
		},
	}
}
