package debug

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/urfave/cli/v3"
	"k8s.io/client-go/kubernetes"
)

const debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 1 {
		return ctx, fmt.Errorf("missing required arguments: %v", cmd.ArgsUsage)
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	cfg := MakeConfig(cmd)
	clientset, err := SetupClient(cfg, cmd)
	if err != nil {
		return err
	}

	dg := Setup(clientset, cfg)
	if err := dg.Debug(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}

func SetupClient(cfg *Config, cmd *cli.Command) (kubernetes.Interface, error) {
	cluster := cmd.String("context")
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

func MakeConfig(cmd *cli.Command) *Config {
	return &Config{
		WorkloadName: cmd.Args().First(),
		Namespace:    cmd.String("namespace"),
		DebugImage:   debugImageDefault,
		CopyPod:      cmd.Bool("copy"),
		ByPod:        cmd.Bool("by-pod"),
	}
}
