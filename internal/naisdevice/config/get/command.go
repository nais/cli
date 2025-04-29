package get

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func Action(ctx context.Context, cmd *cli.Command) error {
	config, err := naisdevice.GetConfiguration(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("AutoConnect:\t%v\n", config.AutoConnect)

	return nil
}
