package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/manifoldco/promptui"

	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

func (d *Debug) debugPod(podName string) error {
	cmd := exec.Command(
		"kubectl",
		"debug",
		"-n", d.cfg.Namespace,
		fmt.Sprintf("pod/%s", podName),
		"-it",
		"--stdin",
		"--tty",
		"--context", d.cfg.Context,
		"--profile=restricted",
		"--image", d.cfg.DebugImage)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	fmt.Printf("Run 'nais debug tidy %s' to clean up debug containers, this will delete pod(s) with debug containers \n", d.cfg.WorkloadName)

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

	prompt := promptui.Select{
		Label: "Select pod to Debug",
		Items: podNames,
	}

	_, podName, err := prompt.Run()
	if err != nil {
		fmt.Printf("prompt failed %v\n", err)
		return err
	}

	if err := d.debugPod(podName); err != nil {
		fmt.Printf("failed to debug pod %s: %v\n", podName, err)
	}

	return nil
}
