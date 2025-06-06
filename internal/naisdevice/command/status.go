package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/device/pkg/pb"
)

type handler struct {
	flags *flag.Status
}

func statuscmd(rootFlags *root.Flags) *cli.Command {
	h := &handler{flags: &flag.Status{Flags: rootFlags}}

	return &cli.Command{
		Name:         "status",
		Short:        "Show the status of your naisdevice.",
		Flags:        h.flags,
		RunFunc:      h.Run,
		ValidateFunc: h.Validate,
	}
}

func (h *handler) Validate(_ context.Context, _ []string) error {
	if !slices.Contains([]string{"", "yaml", "json"}, h.flags.Output) {
		return fmt.Errorf("%v is not an implemented formal", h.flags.Output)
	}

	return nil
}

func (h *handler) Run(ctx context.Context, out cli.Output, _ []string) error {
	agentStatus, err := naisdevice.GetStatus(ctx)
	if err != nil {
		return err
	}

	if agentStatus.GetConnectionState() != pb.AgentState_Connected {
		if h.flags.Quiet {
			return nil
		}
		return fmt.Errorf("not connected to naisdevice")
	}

	if h.flags.Output != "" {
		return naisdevice.PrintFormattedStatus(h.flags.Output, agentStatus, out)
	}

	if h.flags.IsVerbose() {
		naisdevice.PrintVerboseStatus(agentStatus, out)
		return nil
	}

	out.Println(agentStatus.ConnectionState.String())

	return nil
}
