package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/v2/internal/naisdevice"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func get(_ *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "get",
		Title: "Get a naisdevice setting.",
		Args: []naistrix.Argument{
			{Name: "setting"},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
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
