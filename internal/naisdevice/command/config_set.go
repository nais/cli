package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/v2/internal/naisdevice"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func set(_ *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "set",
		Title: "Set a configuration value.",
		Args: []naistrix.Argument{
			{Name: "setting"},
			{Name: "value"},
		},
		AutoCompleteFunc: naisdevice.AutocompleteSet,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			setting := args[0]
			value, err := strconv.ParseBool(args[1])
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
