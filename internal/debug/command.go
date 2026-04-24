package debug

import (
	"fmt"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/naistrix"
	"k8s.io/client-go/kubernetes"
)

const debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

func Run(workloadName string, flags *flag.Debug, out *naistrix.OutputWriter) error {
	clientSet, err := SetupClient(flags.DebugSticky, flags.Environment)
	if err != nil {
		return err
	}

	dg := Setup(clientSet, flags.DebugSticky, workloadName, debugImageDefault, flags.ByPod, out)
	if err := dg.Debug(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}

func SetupClient(flags *flag.DebugSticky, cluster flag.Environment) (kubernetes.Interface, error) {
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(string(flags.Environment)))

	team, err := flags.RequiredTeam()
	if err != nil {
		return nil, err
	}
	client.CurrentNamespace = team

	clientSet, err := k8s.SetupClientGo(string(flags.Environment))
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
