package debugcmd

import (
	"fmt"

	"github.com/nais/cli/pkg/debug"
	"github.com/nais/cli/pkg/k8s"
	"github.com/urfave/cli/v2"
)

const (
	namespaceFlagName = "namespace"
	contextFlagName   = "context"
	debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "Debug an application",
		ArgsUsage: "appname",
		Subcommands: []*cli.Command{
			tidyCommand(),
		},
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
			client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
			if cfg.Namespace == "" {
				cfg.Namespace = client.CurrentNamespace
			}

			clientset, err := k8s.SetupClientGo(cluster)
			if err != nil {
				return err
			}

			dg := debug.Setup(clientset, cfg)
			if err := dg.Debug(); err != nil {
				return fmt.Errorf("debugging instance: %w", err)
			}
			return nil
		},
	}
}

func kubeConfigFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        contextFlagName,
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func makeConfig(cCtx *cli.Context) debug.Config {
	appName := cCtx.Args().First()
	namespace := cCtx.Args().Get(1)

	return debug.Config{
		AppName:    appName,
		Namespace:  namespace,
		DebugImage: debugImageDefault,
	}
}
