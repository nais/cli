package status

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

type handler struct {
	*root.Flags
	Quiet  bool
	Output string
}

func Command(rootFlags *root.Flags) *cli.Command {
	h := &handler{Flags: rootFlags}

	return cli.NewCommand("status", "Show the status of your naisdevice.",
		cli.WithFlag("output", "Output format, can be json or yaml", "o", &h.Output),
		cli.WithFlag("quiet", "Suppress output if not connected", "q", &h.Quiet),
		cli.WithRun(h.Run),
		cli.WithValidate(h.Validate),
	)
}

func (h *handler) Validate(ctx context.Context, _ []string) error {
	if !slices.Contains([]string{"", "yaml", "json"}, h.Output) {
		return fmt.Errorf("%v is not an implemented format", h.Output)
	}

	return nil
}

func (h *handler) Run(ctx context.Context, _ []string) error {
	status, err := GetStatus(ctx)
	if err != nil {
		return err
	}

	if !IsConnected(status) {
		if h.Quiet {
			return nil
		}
		return fmt.Errorf("not connected to naisdevice")
	}

	if h.Output != "" {
		return PrintFormattedStatus(h.Output, status)
	}

	if h.IsVerbose() {
		PrintVerboseStatus(status)
		return nil
	}

	fmt.Println(status.ConnectionState.String())

	return nil
}
