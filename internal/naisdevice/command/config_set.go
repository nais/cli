package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
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
		ValidateFunc:     cli.ValidateExactArgs(2),
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
