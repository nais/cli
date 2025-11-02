package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/naisdevice/command/flag"
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
			allGateways, err := naisdevice.GetGateways(ctx)
			if err != nil {
				return naistrix.Errorf("Unable to list gateways, are you connected to naisdevice?")
			}

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
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, toComplete string) (completions []string, activeHelp string) {
			gateways, err := naisdevice.GetGateways(ctx)
			if err != nil {
				return []string{}, "Unable to list gateways, are you connected to naisdevice?"
			}

			gws := make([]string, len(gateways))
			for i, g := range gateways {
				gws[i] = g.Name
			}

			return gws, "Select a gateway to describe."
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			gateway, err := naisdevice.GetGateway(ctx, args.Get("gateway"))
			if err != nil {
				return naistrix.Errorf("Unable to describe gateway, are you connected to naisdevice?")
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
				out.Printf("Name: <info>%s</info>\n", gateway.Name)
				out.Printf("Connected: <info>%s</info>\n", YesNo(gateway.Healthy))
				out.Printf("Public key: <info>%s</info>\n", gateway.PublicKey)
				out.Printf("Endpoint: <info>%s</info>\n", gateway.Endpoint)
				out.Printf("IPv4: <info>%s</info>\n", gateway.Ipv4)
				out.Printf("IPv6: <info>%s</info>\n", gateway.Ipv6)
				out.Printf("Routes (IPv4): <info>%s</info>\n", strings.Join(gateway.RoutesIPv4, ", "))
				out.Printf("Requires JITA: <info>%s</info>\n", YesNo(gateway.RequiresPrivilegedAccess))
				out.Printf("Access group IDs: <info>%s</info>\n", strings.Join(gateway.AccessGroupIDs, ", "))
				return nil
			}

			return o.Render(gateway)
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
