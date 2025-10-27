package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/naistrix"
)

func set() *naistrix.Command {
	return &naistrix.Command{
		Name:  "set",
		Title: "Set a configuration value.",
		Args: []naistrix.Argument{
			{Name: "setting"},
			{Name: "value"},
		},
		AutoCompleteFunc: naisdevice.AutocompleteSet,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			setting := args.Get("setting")
			value, err := strconv.ParseBool(args.Get("value"))
			if err != nil {
				return fmt.Errorf("invalid bool value: %v", err)
			}

			if err := naisdevice.SetConfig(ctx, setting, value); err != nil {
				return err
			}

			out.Printf("%v has been set to %v\n", setting, value)

			return nil
		},
	}
}
