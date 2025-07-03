package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/device/pkg/pb"
	"github.com/nais/naistrix"
)

func statuscmd(rootFlags *root.Flags) *naistrix.Command {
	flags := &flag.Status{Flags: rootFlags}
	return &naistrix.Command{
		Name:  "status",
		Title: "Show the status of your naisdevice.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
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
				return naisdevice.PrintFormattedStatus(string(flags.Output), agentStatus, out)
			}

			if flags.IsVerbose() {
				naisdevice.PrintVerboseStatus(agentStatus, out)
				return nil
			}

			out.Println(agentStatus.ConnectionState.String())

			return nil
		},
		ValidateFunc: func(context.Context, []string) error {
			if !slices.Contains([]string{"", "yaml", "json"}, string(flags.Output)) {
				return naistrix.Errorf("%v is not an implemented format", flags.Output)
			}

			return nil
		},
	}
}
