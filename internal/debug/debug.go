package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

	pod, err := interactiveSelectPod(pods.Items)
	if err != nil {
		pterm.Error.Printf("Failed to select pod: %v\n", err)
		return err
	}

	if err := d.debugPod(*pod); err != nil {
		pterm.Error.Printf("Failed to debug pod %s: %v\n", pod.Name, err)
	}

	return nil
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

func (d *Debug) podExists(name string) (bool, error) {
	if _, err := d.podsClient.Get(d.ctx, name, metav1.GetOptions{}); err == nil {
		return true, nil
	} else if k8serrors.IsNotFound(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (d *Debug) debugPod(pod corev1.Pod) error {
	args := []string{
		"debug",
		"pod/" + pod.Name,
		"--namespace", d.flags.Namespace,
		"--context", string(d.flags.Context),
		"--stdin",
		"--tty",
		"--profile=restricted",
		"--image", debugImage,
		"--quiet",
	}

	if d.flags.Copy {
		return d.createDebugPod(args, pod)
	}

	return d.createDebugContainer(args, &pod)
}

func (d *Debug) createDebugContainer(commonArgs []string, pod *corev1.Pod) error {
	pterm.Info.Printf("Creating ephemeral debug container in pod %s...\n", pod.Name)

	args := append(commonArgs, "--target", d.workloadName) // workloadName is the same as container name for nais apps
	if err := d.kubectl(d.ctx, true, args...); err != nil {
		return fmt.Errorf("failed to start debug command: %v", err)
	}

	pterm.Info.Println("Remember to restart the pod to remove the debug container")
	return nil
}

func (d *Debug) createDebugPod(commonArgs []string, pod corev1.Pod) error {
	pterm.Info.Printf("Creating a copy of pod %s with a debug container...\n", pod.Name)

	podCopyName := debugPodName(pod.Name)
	// If debug pod already exists, attach instead of creating a new one
	if exists, err := d.podExists(podCopyName); exists {
		pterm.Info.Printf("Debug pod %q already exists, attaching...\n", podCopyName)
		return d.whenDebugContainerReady(podCopyName, d.attach)
	} else if err != nil {
		return fmt.Errorf("failed to check for existing debug pod %q: %v", podCopyName, err)
	}

	args := append(commonArgs,
		"--copy-to", debugPodName(pod.Name),
		"--container", debugPodContainerName,
		"--keep-annotations",
		"--keep-liveness",
		"--keep-readiness",
		"--keep-startup",
		"--attach=false",
	)
	if err := d.kubectl(d.ctx, false, args...); err != nil {
		return fmt.Errorf("failed to start debug command: %v", err)
	}

	if err := d.annotateAndLabelDebugPod(podCopyName, pod.Labels); err != nil {
		return fmt.Errorf("failed to annotate and label debug pod: %w", err)
	}

	if err := d.whenDebugContainerReady(podCopyName, d.attach); err != nil {
		return fmt.Errorf("failed to attach to debug pod %q: %w", podCopyName, err)
	}

	// TODO ask if the user wants to delete the debug pod after attaching
	pterm.Info.Printf("Debug pod will self-destruct in %s\n", d.flags.Ttl)
	return nil
}

func (d *Debug) annotateAndLabelDebugPod(debugPodName string, existingLabels map[string]string) error {
	args := []string{
		"label",
		"pod/" + debugPodName,
		"cli.nais.io/debug=true",
		"euthanaisa.nais.io/enabled=true",
	}

	delete(existingLabels, "pod-template-hash")
	for label, value := range existingLabels {
		args = append(args, fmt.Sprintf("%s=%s", label, value))
	}

	if err := d.kubectl(
		d.ctx,
		false,
		args...,
	); err != nil {
		return fmt.Errorf("unable to label debug pod: %w", err)
	}

	killAfter := time.Now().Add(d.flags.Ttl).Format(time.RFC3339)

	if err := d.kubectl(
		d.ctx,
		false,
		"annotate",
		"pod/"+debugPodName,
		"euthanaisa.nais.io/kill-after="+killAfter,
	); err != nil {
		return fmt.Errorf("unable to annotate debug pod: %w", err)
	}

	return nil
}

func (d *Debug) attach(podName string) error {
	if err := d.kubectl(d.ctx,
		true,
		"attach",
		"pod/"+podName,
		"--container", debugPodContainerName,
		"--stdin",
		"--tty",
		"--quiet",
	); err != nil {
		return fmt.Errorf("failed to attach to the debug container: %w", err)
	}

	pterm.Info.Println("Exited from Debug container")

	return nil
}

func interactiveSelectPod(pods []corev1.Pod) (*corev1.Pod, error) {
	if len(pods) > 1 {
		var podNames []string
		for _, p := range pods {
			podNames = append(podNames, p.Name)
		}

		result, err := pterm.DefaultInteractiveSelect.WithOptions(podNames).Show()
		if err != nil {
			pterm.Error.Printf("Prompt failed: %v\n", err)
			return nil, err
		}

		for _, p := range pods {
			if p.Name == result {
				return &p, nil
			}
		}
	} else if len(pods) == 1 {
		return &pods[0], nil
	}

	return nil, fmt.Errorf("no pod selected or found")
}

func labelSelector(key, value string) metav1.ListOptions {
	excludeDebugPods := "cli.nais.io/debug!=true"
	return metav1.ListOptions{
		LabelSelector: strings.Join([]string{excludeDebugPods, key + "=" + value}, ","),
	}
}

// debugPodName generates a name for the debug pod copy given a pod name.
func debugPodName(podName string) string {
	return podName + "-" + debugPodSuffix
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

func (d *Debug) kubectl(ctx context.Context, attach bool, args ...string) error {
	cmd := exec.CommandContext(ctx,
		"kubectl",
		append(args,
			"--namespace", d.flags.Namespace,
			"--context", string(d.flags.Context),
		)...,
	)

	if d.flags.IsDebug() {
		pterm.Info.Println("Running command:", strings.Join(cmd.Args, " "))
	}

	if attach {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl command failed: %w\nOutput: %s", err, string(out))
	}

	if d.flags.IsVerbose() {
		pterm.Info.Println("Command output:", string(out))
	}

	return nil
}
