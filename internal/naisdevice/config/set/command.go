package set

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 2 {
		metrics.AddOne(ctx, "device_settings_error_total")
		return ctx, fmt.Errorf("missing required arguments: setting, value")
	}

	setting := cmd.Args().Get(0)
	value := cmd.Args().Get(1)
	if !slices.Contains(naisdevice.GetAllowedSettings(true, true), strings.ToLower(setting)) {
		metrics.AddOne(ctx, "device_settings_error_total")
		return ctx, fmt.Errorf("%v is not one of the allowed settings: %v", setting, strings.Join(naisdevice.GetAllowedSettings(false, false), ", "))
	}

	if _, err := strconv.ParseBool(value); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	setting := cmd.Args().Get(0)
	valueString := cmd.Args().Get(1)

	value, err := strconv.ParseBool(valueString)
	if err != nil {
		return err
	}

	if err := naisdevice.SetConfiguration(ctx, setting, value); err != nil {
		return err
	}

	fmt.Printf("%v has been set to %v\n", setting, value)

	return nil
}
