package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func get(_ *root.Flags) *cli.Command {
	return cli.NewCommand("get", "Get a naisdevice setting.",
		cli.WithArgs("setting"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, w output.Output, args []string) error {
			setting := args[0]

			values, err := naisdevice.GetConfig(ctx)
			if err != nil {
				return err
			}

			if strings.EqualFold(setting, "autoconnect") {
				w.Printf("%v:\t%v\n", setting, values.AutoConnect)
			} else if strings.EqualFold(setting, "iloveninetiesboybands") {
				w.Printf("%v:\t%v\n", setting, values.ILoveNinetiesBoybands)
			} else {
				return fmt.Errorf("unknown setting: %v", setting)
			}

			return nil
		}),
	)
}
