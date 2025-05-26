package set

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

func Set(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("set", "Set a configuration value.",
		cli.WithArgs("setting", "value"),
		cli.WithAutoComplete(autocomplete),
		cli.WithRun(run),
		cli.WithValidate(validate),
	)
}

func validate(_ context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected exactly 2 arguments, got %d", len(args))
	}

	return nil
}

func autocomplete(ctx context.Context, args []string, _ string) ([]string, string) {
	if len(args) == 0 {
		return GetAllowedSettings(false, false), ""
	} else if len(args) == 1 {
		var completions []string
		for key, value := range GetSettingValues(args[0]) {
			completions = append(completions, key+"\t"+value)
		}
		return completions, "Possible values"
	}

	return nil, "no more inputs expected, press enter"
}

func run(ctx context.Context, args []string) error {
	setting := args[0]
	value, err := strconv.ParseBool(args[1])
	if err != nil {
		return fmt.Errorf("invalid bool value: %v", err)
	}

	if err := set(ctx, setting, value); err != nil {
		return err
	}

	fmt.Printf("%v has been set to %v\n", setting, value)

	return nil
}
