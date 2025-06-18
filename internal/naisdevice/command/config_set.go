package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
)

func set(_ *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "set",
		Title: "Set a configuration value.",
		Args: []cli.Argument{
			{Name: "setting", Required: true},
			{Name: "value", Required: true},
		},
		AutoCompleteFunc: naisdevice.AutocompleteSet,
		ValidateFunc: func(_ context.Context, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected exactly 2 arguments, got %d", len(args))
			}

			return nil
		},
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
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
