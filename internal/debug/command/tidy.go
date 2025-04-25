package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/debug"
	"github.com/urfave/cli/v3"
)

func tidy() *cli.Command {
	return &cli.Command{
		Name:        "tidy",
		Usage:       "Clean up debug containers and debug pods",
		Description: "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
		ArgsUsage:   "workloadname",
		Flags: []cli.Flag{
			kubeConfigFlag(),
			namespaceFlag(),
			copyFlag(),
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Args().Len() < 1 {
				return ctx, fmt.Errorf("missing required arguments: %v", cmd.ArgsUsage)
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := makeConfig(cmd)
			clientset, err := setupClient(cfg, cmd)
			if err != nil {
				return err
			}

			dg := debug.Setup(clientset, cfg)
			if err := dg.Tidy(); err != nil {
				return fmt.Errorf("debugging instance: %w", err)
			}
			return nil
		},
	}
}
