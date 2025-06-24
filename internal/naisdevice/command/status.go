package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
	"github.com/nais/device/pkg/pb"
)

func statuscmd(rootFlags *root.Flags) *cli.Command {
	flags := &flag.Status{Flags: rootFlags}
	return &cli.Command{
		Name:  "status",
		Title: "Show the status of your naisdevice.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			agentStatus, err := naisdevice.GetStatus(ctx)
			if err != nil {
				return err
			}

			if agentStatus.GetConnectionState() != pb.AgentState_Connected {
				if flags.Quiet {
					return nil
				}
				return fmt.Errorf("not connected to naisdevice")
			}

			if flags.Output != "" {
				return naisdevice.PrintFormattedStatus(flags.Output, agentStatus, out)
			}

			if flags.IsVerbose() {
				naisdevice.PrintVerboseStatus(agentStatus, out)
				return nil
			}

			out.Println(agentStatus.ConnectionState.String())

			return nil
		},
		ValidateFunc: func(context.Context, []string) error {
			if !slices.Contains([]string{"", "yaml", "json"}, flags.Output) {
				return cli.Errorf("%v is not an implemented format", flags.Output)
			}

			return nil
		},
	}
}
