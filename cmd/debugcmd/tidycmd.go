package debugcmd

import (
	"fmt"

	"github.com/nais/cli/pkg/debug"
	"github.com/nais/cli/pkg/k8s"
	"github.com/urfave/cli/v2"
)

func tidyCommand() *cli.Command {
	return &cli.Command{
		Name:      "tidy",
		Usage:     "Clean up ephemeral containers and debug pods",
		ArgsUsage: "appname",
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
			cluster := cCtx.String(contextFlagName)
			namespace := cCtx.String(namespaceFlagName)
			client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace
			if namespace != "" {
				cfg.Namespace = namespace
			}

			clientset, err := k8s.SetupClientGo(cluster)
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
