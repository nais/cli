package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/task"
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
	pods, err := task.Timed(d.ctx, d.flags.Timeout, "Fetching pods for workload", func(ctx context.Context) (*corev1.PodList, error) {
		return d.getPodsForWorkload(ctx)
	})
	if err != nil {
		pterm.Error.Println("Failed to get pods for workload")
		return err
	}

	if len(pods.Items) == 0 {
		pterm.Info.Println("No pods found.")
		return nil
	}

	pod, err := interactiveSelectPod(pods.Items)
	if err != nil {
		pterm.Error.Println("Failed to select pod")
		return err
	}

	if err := d.debugPod(*pod); err != nil {
		pterm.Error.Println("Failed to debug pod")
		return err
	}

	return nil
}

func (d *Debug) getPodsForWorkload(ctx context.Context) (*corev1.PodList, error) {
	podList, err := d.podsClient.List(ctx, labelSelector("app.kubernetes.io/name", d.workloadName))
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	if len(podList.Items) > 0 {
		return podList, nil
	}

	podList, err = d.podsClient.List(ctx, labelSelector("app=", d.workloadName))
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	return podList, nil
}

func (d *Debug) podExists(ctx context.Context, name string) (bool, error) {
	if _, err := d.podsClient.Get(ctx, name, metav1.GetOptions{}); err == nil {
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

	return d.createDebugContainer(args)
}

func (d *Debug) createDebugContainer(commonArgs []string) error {
	args := append(commonArgs, "--target", d.workloadName) // workloadName is the same as container name for nais apps

	_, err := task.Timed(d.ctx, d.flags.Timeout, "Creating ephemeral debug container", func(ctx context.Context) (*any, error) {
		return nil, d.kubectl(ctx, true, args...)
	})
	if err != nil {
		pterm.Error.Println("Failed to create ephemeral debug container")
		return err
	}

	pterm.Info.Println("Remember to restart the pod to remove the debug container")
	return nil
}

func (d *Debug) createDebugPod(commonArgs []string, pod corev1.Pod) error {
	debugPodName := createDebugPodName(pod.Name)

	exists, err := task.Timed(d.ctx, d.flags.Timeout, "Check for existing debug pod", func(ctx context.Context) (bool, error) {
		return d.podExists(ctx, debugPodName)
	})
	if err != nil {
		return fmt.Errorf("failed to check for existing debug pod: %w", err)
	} else if exists {
		return d.attach(d.ctx, debugPodName)
	}

	args := append(commonArgs,
		"--copy-to", debugPodName,
		"--container", debugPodContainerName,
		"--keep-annotations",
		"--keep-liveness",
		"--keep-readiness",
		"--keep-startup",
		"--attach=false",
	)
	_, err = task.Timed(d.ctx, d.flags.Timeout, "Create debug pod", func(ctx context.Context) (*any, error) {
		return nil, d.kubectl(ctx, false, args...)
	})
	if err != nil {
		return fmt.Errorf("failed to create debug pod: %v", err)
	}

	_, err = task.Timed(d.ctx, d.flags.Timeout, "Annotate debug pod", func(ctx context.Context) (*any, error) {
		return nil, d.annotateAndLabelDebugPod(ctx, debugPodName, pod.Labels)
	})
	if err != nil {
		return fmt.Errorf("failed to annotate and label debug pod: %w", err)
	}

	if err := d.attach(d.ctx, debugPodName); err != nil {
		return fmt.Errorf("failed to attach to debug pod %q: %w", debugPodName, err)
	}

	// TODO ask if the user wants to delete the debug pod after attaching
	pterm.Info.Printf("Debug pod will self-destruct in %s\n", d.flags.TTL)
	return nil
}

func (d *Debug) annotateAndLabelDebugPod(ctx context.Context, debugPodName string, existingLabels map[string]string) error {
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
		ctx,
		false,
		args...,
	); err != nil {
		return fmt.Errorf("unable to label debug pod: %w", err)
	}

	killAfter := time.Now().Add(d.flags.TTL).Format(time.RFC3339)

	if err := d.kubectl(
		ctx,
		false,
		"annotate",
		"pod/"+debugPodName,
		"euthanaisa.nais.io/kill-after="+killAfter,
	); err != nil {
		return fmt.Errorf("unable to annotate debug pod: %w", err)
	}

	return nil
}

func interactiveSelectPod(pods []corev1.Pod) (*corev1.Pod, error) {
	if len(pods) > 1 {
		var podNames []string
		for _, p := range pods {
			podNames = append(podNames, p.Name)
		}

		result, err := pterm.DefaultInteractiveSelect.WithOptions(podNames).WithDefaultText(pterm.Normal("Please select a pod")).Show()
		if err != nil {
			pterm.Error.Println("Prompt failed")
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
func createDebugPodName(podName string) string {
	return podName + "-" + debugPodSuffix
}

func (d *Debug) debugContainerIsReady(podName string) func(ctx context.Context) (*corev1.Pod, error) {
	return func(ctx context.Context) (*corev1.Pod, error) {
		pod, err := d.podsClient.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		for _, c := range pod.Status.ContainerStatuses {
			if c.Name == debugPodContainerName && c.State.Running != nil {
				return pod, nil
			}
		}

		return nil, fmt.Errorf("no ready debug container with name %q found in pod %q", debugPodContainerName, podName)
	}
}

func (d *Debug) attach(ctx context.Context, podName string) error {
	_, err := task.Timed(ctx, d.flags.Timeout, "Attaching to container", func(ctx context.Context) (*any, error) {
		_, err := withRetryOnErr(d.debugContainerIsReady(podName))(ctx)
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("debug container did not start: %w", err)
	}

	pterm.Info.Printf("You are now typing in the debug container in %q. Type exit to exit.\n", podName)
	return d.kubectl(ctx, true, "attach", "pod/"+podName, "--container", debugPodContainerName, "--stdin", "--tty", "--quiet")
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

// withRetryOnErr retries the function until it returns nil error, or context is done.
func withRetryOnErr[T any](f func(context.Context) (*T, error)) func(context.Context) (*T, error) {
	return func(ctx context.Context) (*T, error) {
		ret, err := f(ctx)
		for err != nil {
			ret, err = f(ctx)
		}

		return ret, err
	}
}
