package debugcmd

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/k8s"
	"github.com/urfave/cli/v2"
)

const (
	contextFlagName   = "context"
	copyFlagName      = "copy"
	namespaceFlagName = "namespace"
	byPodFlagName     = "by-pod"
	debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "Create and attach to a debug container",
		ArgsUsage: "workloadname",
		Description: "Create and attach to a debug pod or container. \n" +
			"When flag '--copy' is set, the command can be used to debug a copy of the original pod, \n" +
			"allowing you to troubleshoot without affecting the live pod.\n" +
			"To debug a live pod, run the command without the '--copy' flag.\n" +
			"You can only reconnect to the debug session if the pod is running.",
		Subcommands: []*cli.Command{
			tidyCommand(),
		},
		Flags: []cli.Flag{
			kubeConfigFlag(),
			copyFlag(),
			namespaceFlag(),
			byPodFlag(),
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
			if err := dg.Debug(); err != nil {
				return fmt.Errorf("debugging instance: %w", err)
			}
			return nil
		},
	}
}

func setupClient(cfg *debug.Config, cCtx *cli.Context) (kubernetes.Interface, error) {
	cluster := cCtx.String(contextFlagName)
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))

	if cfg.Namespace == "" {
		cfg.Namespace = client.CurrentNamespace
	}
	if cluster != "" {
		cfg.Context = cluster
	}

	clientset, err := k8s.SetupClientGo(cluster)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func kubeConfigFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        contextFlagName,
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func byPodFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "by-pod",
		Aliases:     []string{"p"},
		Usage:       "Attach to a specific `BY-POD` in a workload",
		DefaultText: "The first pod in the workload",
	}
}

func copyFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        copyFlagName,
		Aliases:     []string{"cp"},
		Usage:       "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected",
		DefaultText: "Attach to the current 'live' pod",
	}
}

func namespaceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        namespaceFlagName,
		Aliases:     []string{"n"},
		Usage:       "The `NAMESPACE` to use",
		DefaultText: "The current namespace in your kubeconfig",
	}
}

func makeConfig(cCtx *cli.Context) *debug.Config {
	appName := cCtx.Args().First()

	return &debug.Config{
		WorkloadName: appName,
		Namespace:    cCtx.String(namespaceFlagName),
		DebugImage:   debugImageDefault,
		CopyPod:      cCtx.Bool(copyFlagName),
		ByPod:        cCtx.Bool(byPodFlagName),
	}
}
