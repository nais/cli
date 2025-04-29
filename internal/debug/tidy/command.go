package tidy

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/debug"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 1 {
		return ctx, fmt.Errorf("missing required arguments: %v", cmd.ArgsUsage)
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	cfg := debug.MakeConfig(cmd)
	clientset, err := debug.SetupClient(cfg, cmd)
	if err != nil {
		return err
	}

	dg := debug.Setup(clientset, cfg)
	if err := dg.Tidy(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}
