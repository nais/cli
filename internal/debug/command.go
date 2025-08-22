package debug

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"k8s.io/client-go/kubernetes"
)

func Run(ctx context.Context, workloadName string, flags *flag.Debug) error {
	clientSet, err := SetupClient(flags, flags.Context)
	if err != nil {
		return err
	}

	dg := &Debug{
		podsClient:   clientSet.CoreV1().Pods(flags.Namespace),
		flags:        flags,
		workloadName: workloadName,
	}
	if err := dg.Debug(ctx); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}

func SetupClient(flags *flag.Debug, cluster flag.Context) (kubernetes.Interface, error) {
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(string(cluster)))

	if flags.Namespace == "" {
		flags.Namespace = client.CurrentNamespace
	}

	if cluster != "" {
		flags.Context = cluster
	}

	clientSet, err := k8s.SetupClientGo(string(cluster))
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
