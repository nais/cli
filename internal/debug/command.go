package debug

import (
	"fmt"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"k8s.io/client-go/kubernetes"
)

const debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

func Run(workloadName string, flags *flag.Debug) error {
	cfg := MakeConfig(workloadName, flags)
	clientSet, err := SetupClient(cfg, flags.Context)
	if err != nil {
		return err
	}

	dg := Setup(clientSet, cfg)
	if err := dg.Debug(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}

func SetupClient(cfg *Config, cluster string) (kubernetes.Interface, error) {
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))

	if cfg.Namespace == "" {
		cfg.Namespace = client.CurrentNamespace
	}

	if cluster != "" {
		cfg.Context = cluster
	}

	clientSet, err := k8s.SetupClientGo(cluster)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func MakeConfig(workloadName string, flags *flag.Debug) *Config {
	return &Config{
		WorkloadName: workloadName,
		Namespace:    flags.Namespace,
		DebugImage:   debugImageDefault,
		CopyPod:      flags.Copy,
		ByPod:        flags.ByPod,
	}
}
