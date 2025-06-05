package debug

import (
	"fmt"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"k8s.io/client-go/kubernetes"
)

const debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

func Run(workloadName string, flags *flag.Debug) error {
	clientSet, err := SetupClient(flags.DebugSticky, flags.Context)
	if err != nil {
		return err
	}

	dg := Setup(clientSet, flags.DebugSticky, workloadName, debugImageDefault, flags.ByPod)
	if err := dg.Debug(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}

func SetupClient(flags *flag.DebugSticky, cluster string) (kubernetes.Interface, error) {
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))

	if flags.Namespace == "" {
		flags.Namespace = client.CurrentNamespace
	}

	if cluster != "" {
		flags.Context = cluster
	}

	clientSet, err := k8s.SetupClientGo(cluster)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
