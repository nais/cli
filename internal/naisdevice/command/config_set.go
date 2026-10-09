package command

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/naistrix"
)

func set() *naistrix.Command {
	return &naistrix.Command{
		Name:        "set",
		Title:       "Set a configuration value.",
		Description: "Set a naisdevice configuration value. The setting name and a boolean value (true/false) are required.",
		Args: []naistrix.Argument{
			{Name: "setting"},
			{Name: "value", Choices: []string{"true", "false"}, ChoicesCaseInsensitive: true},
		},
		AutoCompleteFunc: naisdevice.AutocompleteSet,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			setting := args.Get("setting")
			value := args.Get("value") == "true"

			if err := naisdevice.SetConfig(ctx, setting, value); err != nil {
				return err
			}

			out.Printf("%v has been set to %v\n", setting, value)

			return nil
		},
	}
}
