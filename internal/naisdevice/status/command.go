package status

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

type Flags struct {
	Quiet   bool
	Verbose bool
	Output  string
}

func Action(ctx context.Context, flags Flags) error {
	status, err := naisdevice.GetStatus(ctx)
	if err != nil {
		return err
	}

	if flags.Quiet {
		if !naisdevice.IsConnected(status) {
			return cli.Exit("", 1)
		}
		return nil
	}

	if flags.Output != "" {
		return naisdevice.PrintFormattedStatus(flags.Output, status)
	}

	if flags.Verbose {
		naisdevice.PrintVerboseStatus(status)
		return nil
	}

	fmt.Println(status.ConnectionState.String())

	return nil
}
