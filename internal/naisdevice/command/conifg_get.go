package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/naistrix"
)

func get() *naistrix.Command {
	return &naistrix.Command{
		Name:  "get",
		Title: "Get a naisdevice setting.",
		Args: []naistrix.Argument{
			{Name: "setting"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			setting := args.Get("setting")

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
