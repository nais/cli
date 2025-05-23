package status

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

type Flags struct {
	*root.Flags
	Quiet  bool
	Output string
}

func Status(rootFlags *root.Flags) *cli.Command {
	flags := &Flags{Flags: rootFlags}

	return cli.NewCommand("status", "Show the status of your naisdevice.",
		cli.WithFlag("output", "Output format, can be json or yaml", "o", &flags.Output),
		cli.WithFlag("quiet", "Suppress output if not connected", "q", &flags.Quiet),
		cli.WithHandler(run(flags)),
	)
}

func run(flags *Flags) cli.HandlerFunc {
	return func(ctx context.Context) error {
		status, err := GetStatus(ctx)
		if err != nil {
			return err
		}

		if !IsConnected(status) {
			if flags.Quiet {
				return nil
			}
			return fmt.Errorf("not connected to naisdevice")
		}

		if flags.Output != "" {
			return PrintFormattedStatus(flags.Output, status)
		}

		if flags.IsVerbose() {
			PrintVerboseStatus(status)
			return nil
		}

		fmt.Println(status.ConnectionState.String())

		return nil
	}
}
