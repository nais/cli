package status

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
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
		cli.WithFlag("output", "o", "Output format, can be json or yaml", &h.Output, cli.FlagRequired()),
		cli.WithFlag("quiet", "q", "Suppress output if not connected", &h.Quiet),
		cli.WithRun(h.Run),
		cli.WithValidate(h.Validate),
	)
}

func (h *handler) Validate(_ context.Context, _ []string) error {
	if !slices.Contains([]string{"", "yaml", "json"}, h.Output) {
		return fmt.Errorf("%v is not an implemented format", h.Output)
	}

	return nil
}

func (h *handler) Run(ctx context.Context, w output.Output, _ []string) error {
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

	w.Println(status.ConnectionState.String())

	return nil
}
