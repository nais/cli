package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/manifoldco/promptui"

	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	debuggerSuffix = "nais-debugger"
)

type Debug struct {
	ctx    context.Context
	client kubernetes.Interface
	cfg    *Config
}

type Config struct {
	Namespace    string
	Context      string
	WorkloadName string
	DebugImage   string
	CopyPod      bool
	ByPod        bool
}

func Setup(client kubernetes.Interface, cfg *Config) *Debug {
	return &Debug{
		ctx:    context.Background(),
		client: client,
		cfg:    cfg,
	}
}

func (d *Debug) getPodsForWorkload() (*core_v1.PodList, error) {
	var podList *core_v1.PodList
	var err error
	podList, err = d.client.CoreV1().Pods(d.cfg.Namespace).List(d.ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", d.cfg.WorkloadName),
	})
	if len(podList.Items) == 0 {
		podList, err = d.client.CoreV1().Pods(d.cfg.Namespace).List(d.ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", d.cfg.WorkloadName),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}
	return podList, nil
}

func debuggerContainerName(podName string) string {
	return fmt.Sprintf("%s-%s", podName, debuggerSuffix)
}

func (d *Debug) debugPod(podName string) error {
	if d.cfg.CopyPod {
		pN := debuggerContainerName(podName)
		_, err := d.client.CoreV1().Pods(d.cfg.Namespace).Get(d.ctx, pN, metav1.GetOptions{})
		if err == nil {
			fmt.Printf("Debug pod copy %s already exists. Attaching...\n", pN)
			// Debug pod copy already exists, attach to it
			return d.attachToExistingDebugContainer(pN)
		} else if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for existing debug pod copy %s: %v", pN, err)
		}
	} else {
		pod, err := d.client.CoreV1().Pods(d.cfg.Namespace).Get(d.ctx, podName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod %s: %v", podName, err)
		}

		if len(pod.Spec.EphemeralContainers) > 0 {
			fmt.Printf("The container %s already has %d terminated debug containers. \n", podName, len(pod.Spec.EphemeralContainers))
			fmt.Printf("Please consider using 'nais debug tidy %s' to clean up\n", d.cfg.WorkloadName)
		}
	}

	return d.createDebugPod(podName)
}

func (d *Debug) attachToExistingDebugContainer(podName string) error {
	defaultDebuggerName := "debugger"
	cmd := exec.Command(
		"kubectl",
		"attach",
		"-n", d.cfg.Namespace,
		fmt.Sprintf("pod/%s", podName),
		"-c", defaultDebuggerName,
		"-i",
		"-t",
		"--context", d.cfg.Context,
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start attach command: %v", err)
	}
	fmt.Printf("Attaching to existing debug container %s in pod %s\n", defaultDebuggerName, podName)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("attach command failed: %v", err)
	}

	return nil
}

func (d *Debug) createDebugPod(podName string) error {
	args := []string{
		"debug",
		"-n", d.cfg.Namespace,
		fmt.Sprintf("pod/%s", podName),
		"-it",
		"--stdin",
		"--tty",
		"--context", d.cfg.Context,
		"--profile=restricted",
		"-q",
		"--image", d.cfg.DebugImage,
	}

	if d.cfg.CopyPod {
		args = append(args,
			"--copy-to", debuggerContainerName(podName),
			"-c", "debugger",
		)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start debug command: %v", err)
	}

	if d.cfg.CopyPod {
		fmt.Printf("Debugging pod copy created, enable process namespace sharing in %s\n", debuggerContainerName(podName))
	} else {
		fmt.Printf("Debugging container created...\n")
	}
	fmt.Printf("Using debugger image %s\n", d.cfg.DebugImage)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("debug command failed: %v", err)
	}

	if d.cfg.CopyPod {
		fmt.Printf("Run 'nais debug -cp %s' command to attach to the debug pod\n", podName)
	}

	return nil
}

func (d *Debug) Debug() error {
	pods, err := d.getPodsForWorkload()
	if err != nil {
		return err
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	if len(podNames) == 0 {
		fmt.Println("No pods found.")
		return nil
	}

	podName := podNames[0]
	if d.cfg.ByPod {
		prompt := promptui.Select{
			Label: "Select pod to Debug",
			Items: podNames,
		}

		_, podName, err = prompt.Run()
		if err != nil {
			fmt.Printf("prompt failed %v\n", err)
			return err
		}
	}

	if err := d.debugPod(podName); err != nil {
		fmt.Printf("failed to debug pod %s: %v\n", podName, err)
	}

	return nil
}
