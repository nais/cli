package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// debugImage is the image used for the debug container.
	debugImage = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

	// debugPodSuffix will be appended to the pod name when creating a debug pod.
	debugPodSuffix = "nais-debugger"

	// debugPodContainerName is the name of the container that will be created in the debug pod. This name is not used
	// when creating ephemeral debug containers.
	debugPodContainerName = "debugger"
)

type Debug struct {
	ctx          context.Context
	podsClient   v1.PodInterface
	flags        *flag.Debug
	workloadName string
}

func (d *Debug) Debug() error {
	pods, err := d.getPodsForWorkload()
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		pterm.Info.Println("No pods found.")
		return nil
	}

	podNames := make([]string, len(pods.Items))
	for i, pod := range pods.Items {
		podNames[i] = pod.Name
	}

	podName := podNames[0]
	if len(podNames) > 1 {
		result, err := pterm.DefaultInteractiveSelect.WithOptions(podNames).Show()
		if err != nil {
			pterm.Error.Printf("Prompt failed: %v\n", err)
			return err
		}
		podName = result
	}

	if err := d.debugPod(podName); err != nil {
		pterm.Error.Printf("Failed to debug pod %s: %v\n", podName, err)
	}

	return nil
}

func (d *Debug) getPodsForWorkload() (*corev1.PodList, error) {
	pterm.Info.Println("Fetching pods for workload...")
	podList, err := d.podsClient.List(d.ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=" + d.workloadName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	if len(podList.Items) > 0 {
		return podList, nil
	}

	podList, err = d.podsClient.List(d.ctx, metav1.ListOptions{
		LabelSelector: "app=" + d.workloadName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	return podList, nil
}

func (d *Debug) debugPod(podName string) error {
	if d.flags.Copy {
		podCopyName := debugPodName(podName)
		if _, err := d.podsClient.Get(d.ctx, podCopyName, metav1.GetOptions{}); err == nil {
			maxRetries := 5
			pollInterval := 2

			pterm.Info.Println("Found existing debug pod, trying to attach...")
			for i := range maxRetries {
				pterm.Info.Printf("Attempt %d/%d\n", i+1, maxRetries)

				pod, err := d.podsClient.Get(d.ctx, podCopyName, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("failed to get debug copy %q: %v", podCopyName, err)
				}

				for _, c := range pod.Status.ContainerStatuses {
					if c.Name == debugPodContainerName && c.State.Running != nil {
						return d.attachToExistingDebugContainer(podCopyName)
					}
				}

				time.Sleep(time.Duration(pollInterval) * time.Second)
			}

			return fmt.Errorf("unable to attach to the existing debug pod")
		} else if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for existing debug pod %q: %v", podCopyName, err)
		}
	}

	return d.createDebugPod(podName)
}

func (d *Debug) createDebugPod(podName string) error {
	args := []string{
		"debug",
		"pod/" + podName,
		"--namespace", d.flags.Namespace,
		"--context", string(d.flags.Context),
		"--stdin",
		"--tty",
		"--profile=restricted",
		"--image", debugImage,
		"--quiet",
	}

	if d.flags.Copy {
		args = append(args,
			"--copy-to", debugPodName(podName),
			"--container", debugPodContainerName,
		)
	} else {
		args = append(args, "--target", d.workloadName)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	pterm.Info.Println("Starting debug container...")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start debug command: %v", err)
	}

	if d.flags.Copy {
		if err := d.annotateAndLabelDebugPod(debugPodName(podName)); err != nil {
			return fmt.Errorf("failed to annotate and label debug pod: %w", err)
		}

		pterm.Info.Printf("Debug pod will self-destruct in %s\n", d.flags.Ttl)
	} else {
		pterm.Info.Println("Remember to restart the pod to remove the debug container")
	}

	return nil
}

func (d *Debug) annotateAndLabelDebugPod(debugPodName string) error {
	labelCommand := exec.CommandContext(
		d.ctx,
		"kubectl",
		"label",
		"pod/"+debugPodName,
		"euthanaisa.nais.io/enabled=true",
		"--namespace", d.flags.Namespace,
		"--context", string(d.flags.Context),
	)

	if err := labelCommand.Run(); err != nil {
		return fmt.Errorf("unable to label debug pod: %w", err)
	}

	killAfter := time.Now().Add(d.flags.Ttl).Format(time.RFC3339)
	annotateCommand := exec.CommandContext(
		d.ctx,
		"kubectl",
		"annotate",
		"pod/"+debugPodName,
		"euthanaisa.nais.io/kill-after="+killAfter,
		"--namespace", d.flags.Namespace,
		"--context", string(d.flags.Context),
	)

	if err := annotateCommand.Run(); err != nil {
		return fmt.Errorf("unable to annotate debug pod: %w", err)
	}

	return nil
}

func (d *Debug) attachToExistingDebugContainer(podName string) error {
	cmd := exec.Command(
		"kubectl",
		"attach",
		"pod/"+podName,
		"--namespace", d.flags.Namespace,
		"--context", string(d.flags.Context),
		"--container", debugPodContainerName,
		"--stdin",
		"--tty",
		"--quiet",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach to the debug container: %w", err)
	}

	pterm.Info.Println("Exited from Debug container")

	return nil
}

// debugPodName generates a name for the debug pod copy given a pod name.
func debugPodName(podName string) string {
	return podName + "-" + debugPodSuffix
}
