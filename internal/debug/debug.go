package debug

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
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

	podMap := make(map[string]corev1.Pod)
	for _, pod := range pods.Items {
		podMap[pod.Name] = pod
	}

	podNames := slices.Collect(maps.Keys(podMap))
	podName := podNames[0]
	if len(podMap) > 1 {
		result, err := pterm.DefaultInteractiveSelect.WithOptions(podNames).Show()
		if err != nil {
			pterm.Error.Printf("Prompt failed: %v\n", err)
			return err
		}
		podName = result
	}

	if err := d.debugPod(podName, podMap[podName].Labels); err != nil {
		pterm.Error.Printf("Failed to debug pod %s: %v\n", podName, err)
	}

	return nil
}

func labelSelector(key, value string) metav1.ListOptions {
	excludeDebugPods := "cli.nais.io/debug!=true"
	return metav1.ListOptions{
		LabelSelector: strings.Join([]string{excludeDebugPods, key + "=" + value}, ","),
	}
}

func (d *Debug) getPodsForWorkload() (*corev1.PodList, error) {
	pterm.Info.Println("Fetching pods for workload...")
	podList, err := d.podsClient.List(d.ctx, labelSelector("app.kubernetes.io/name", d.workloadName))
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	if len(podList.Items) > 0 {
		return podList, nil
	}

	podList, err = d.podsClient.List(d.ctx, labelSelector("app=", d.workloadName))
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	return podList, nil
}

func (d *Debug) whenDebugContainerReady(podCopyName string, callback func(podCopyName string) error) error {
	timeout := 30 * time.Second
	graceDuration := 300 * time.Millisecond

	ctx, cancel := context.WithTimeout(d.ctx, timeout)
	defer cancel()

	statusLine := func(status any) {
		pterm.Printo(fmt.Sprintf("Waiting for debug container to start [%v]", status))
	}

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for debug container to start")
		default:
			deadline, _ := ctx.Deadline()
			statusLine(time.Until(deadline).Round(time.Second))

			pod, err := d.podsClient.Get(d.ctx, podCopyName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get debug copy %q: %v", podCopyName, err)
			}

			for _, c := range pod.Status.ContainerStatuses {
				if c.Name == debugPodContainerName && c.State.Running != nil {
					statusLine(pterm.Green("done"))
					pterm.Println()
					return callback(podCopyName)
				}
			}

			// No running container found, wait a bit before checking again
			time.Sleep(graceDuration)
		}
	}

	return fmt.Errorf("debug pod %q did not start within the expected time", podCopyName)
}

func (d *Debug) podExists(name string) (bool, error) {
	if _, err := d.podsClient.Get(d.ctx, name, metav1.GetOptions{}); err == nil {
		return true, nil
	} else if k8serrors.IsNotFound(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (d *Debug) debugPod(podName string, labels map[string]string) error {
	if d.flags.Copy {
		podCopyName := debugPodName(podName)
		// If debug pod already exists, attach instead of creating a new one
		if exists, err := d.podExists(podCopyName); exists {
			pterm.Info.Printf("Debug pod %q already exists, attaching...\n", podCopyName)
			return d.whenDebugContainerReady(podCopyName, d.attach)
		} else if err != nil {
			return fmt.Errorf("failed to check for existing debug pod %q: %v", podCopyName, err)
		}
	}

	return d.createDebugPod(podName, labels)
}

func (d *Debug) createDebugPod(podName string, labels map[string]string) error {
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
			"--keep-annotations",
			"--keep-liveness",
			"--keep-readiness",
			"--keep-startup",
			"--attach=false",
		)
	} else {
		args = append(args, "--target", d.workloadName)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if d.flags.IsDebug() {
		pterm.Info.Println("Starting debug container: ", cmd.String())
	} else {
		pterm.Info.Println("Starting debug container...")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start debug command: %v", err)
	}

	if d.flags.Copy {
		podCopyName := debugPodName(podName)
		if err := d.annotateAndLabelDebugPod(podCopyName, labels); err != nil {
			return fmt.Errorf("failed to annotate and label debug pod: %w", err)
		}

		if err := d.whenDebugContainerReady(podCopyName, d.attach); err != nil {
			return fmt.Errorf("failed to attach to debug pod %q: %w", podCopyName, err)
		}

		// TODO ask if the user wants to delete the debug pod after attaching
		pterm.Info.Printf("Debug pod will self-destruct in %s\n", d.flags.Ttl)
	} else {
		// TODO ask if the user wants to delete the debug pod after attaching
		pterm.Info.Println("Remember to restart the pod to remove the debug container")
	}

	return nil
}

func (d *Debug) annotateAndLabelDebugPod(debugPodName string, existingLabels map[string]string) error {
	args := []string{
		"--context", string(d.flags.Context),
		"--namespace", d.flags.Namespace,
		"label",
		"pod/" + debugPodName,
		"cli.nais.io/debug=true",
		"euthanaisa.nais.io/enabled=true",
	}
	delete(existingLabels, "pod-template-hash")
	for label, value := range existingLabels {
		args = append(args, fmt.Sprintf("%s=%s", label, value))
	}

	labelCommand := exec.CommandContext(
		d.ctx,
		"kubectl",
		args...,
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

func (d *Debug) attach(podName string) error {
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
