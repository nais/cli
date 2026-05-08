package debug

import (
	"fmt"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/naistrix"
	"k8s.io/client-go/kubernetes"
)

const debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

func Run(workloadName, team, environment string, flags *flag.Debug, out *naistrix.OutputWriter) error {
	clientSet, err := SetupClient(team, environment)
	if err != nil {
		return err
	}

	dg := Setup(clientSet, flags.DebugSticky, workloadName, debugImageDefault, flags.ByPod, out)
	if err := dg.Debug(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}

func SetupClient(team, environment string) (kubernetes.Interface, error) {
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(environment))
	client.CurrentNamespace = team
	clientSet, err := k8s.SetupClientGo(environment)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
