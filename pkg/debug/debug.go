package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"

	v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Debug struct {
	ctx    context.Context
	client kubernetes.Interface
	cfg    Config
}

type Config struct {
	Namespace  string
	Context    string
	AppName    string
	DebugImage string
}

func Setup(client kubernetes.Interface, cfg Config) *Debug {
	return &Debug{
		ctx:    context.Background(),
		client: client,
		cfg:    cfg,
	}
}

func (d *Debug) getApp() (*v1.Deployment, error) {
	app, err := d.client.AppsV1().Deployments(d.cfg.Namespace).Get(d.ctx, d.cfg.AppName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get application in namespace \"%s\": %w", d.cfg.Namespace, err)
	}
	return app, nil
}

func (d *Debug) getPods(app *v1.Deployment) (*core_v1.PodList, error) {
	var podList *core_v1.PodList
	var err error
	podList, err = d.client.CoreV1().Pods(d.cfg.Namespace).List(d.ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", app.Name),
	})
	if len(podList.Items) == 0 {
		podList, err = d.client.CoreV1().Pods(d.cfg.Namespace).List(d.ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", app.Name),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}
	return podList, nil
}

func (d *Debug) Debug() error {
	app, err := d.getApp()
	if err != nil {
		return err
	}

	pods, err := d.getPods(app)
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

	return nil
}

func (d *Debug) Tidy() error {
	app, err := d.getApp()
	if err != nil {
		return err
	}

	pods, err := d.getPods(app)
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

	deleted := 0
	for _, pod := range pods.Items {
		if len(pod.Spec.EphemeralContainers) == 0 {
			continue
		}

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Do you want to delete pod %s", pod.Name),
			IsConfirm: true,
		}

		answer, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrAbort {
				fmt.Printf("Skipping deletion for pod: %s\n", pod.Name)
				continue
			}
			fmt.Printf("Error reading input for pod %s: %v\n", pod.Name, err)
			return err
		}

		// Delete pod if user confirms with "y" or "yes"
		if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
			if err := d.client.CoreV1().Pods(d.cfg.Namespace).Delete(d.ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				fmt.Printf("Failed to delete pod %s: %v\n", pod.Name, err)
			} else {
				deleted++
				fmt.Println("Deleted pod:", pod.Name)
			}
		} else {
			fmt.Println("Skipped pod:", pod.Name)
		}
	}

	if deleted == 0 {
		fmt.Println("No pods with ephemeral containers found.")
	}
	return nil
}
