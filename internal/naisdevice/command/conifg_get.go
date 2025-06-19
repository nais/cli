package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
)

func get(_ *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Title: "Get a naisdevice setting.",
		Args: []cli.Argument{
			{Name: "setting"},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			setting := args[0]

			values, err := naisdevice.GetConfig(ctx)
			if err != nil {
				return err
			}

			if strings.EqualFold(setting, "autoconnect") {
				out.Printf("%v:\t%v\n", setting, values.AutoConnect)
			} else if strings.EqualFold(setting, "iloveninetiesboybands") {
				out.Printf("%v:\t%v\n", setting, values.ILoveNinetiesBoybands)
			} else {
				return fmt.Errorf("unknown setting: %v", setting)
			}

			return nil
		},
	}
}
