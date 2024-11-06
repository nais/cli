package debugcmd

import (
	"fmt"

	"github.com/nais/cli/pkg/debug"
	"github.com/urfave/cli/v2"
)

func tidyCommand() *cli.Command {
	return &cli.Command{
		Name:        "tidy",
		Usage:       "Clean up debug containers and debug pods from your workload",
		Description: "Remove debug containers created by the debug command, the pods will be deleted automatically",
		ArgsUsage:   "workloadname [namespace]",
		Flags: []cli.Flag{
			kubeConfigFlag(),
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("missing required arguments: %v", context.Command.ArgsUsage)
			}

			return nil
		},
		Action: func(cCtx *cli.Context) error {
			cfg := makeConfig(cCtx)
			clientset, err := setupClient(cfg, cCtx)
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
