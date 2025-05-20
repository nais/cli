package status

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisdevice"
)

type Flags struct {
	*naisdevice.Flags
	Quiet  bool
	Output string
}

func Run(ctx context.Context, flags *Flags) error {
	status, err := naisdevice.GetStatus(ctx)
	if err != nil {
		return err
	}

	if !naisdevice.IsConnected(status) {
		if flags.Quiet {
			return nil
		}
		return fmt.Errorf("not connected to naisdevice")
	}

	if flags.Output != "" {
		return naisdevice.PrintFormattedStatus(flags.Output, status)
	}

	if flags.IsVerbose() {
		naisdevice.PrintVerboseStatus(status)
		return nil
	}

	fmt.Println(status.ConnectionState.String())

	return nil
}
