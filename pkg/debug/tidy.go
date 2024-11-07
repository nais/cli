package debug

import (
	"errors"
	"fmt"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/manifoldco/promptui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Debug) Tidy() error {
	pods, err := d.getPodsForWorkload()
	if err != nil {
		return err
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	if len(podNames) == 0 {
		fmt.Println("No pods found")
		return nil
	}

	for _, pod := range pods.Items {
		podName := pod.Name
		if d.cfg.CopyPod {
			podName = debuggerContainerName(pod.Name)
		}

		if !d.cfg.CopyPod && len(pod.Spec.EphemeralContainers) == 0 {
			fmt.Printf("no debug container found for: %s\n", pod.Name)
			continue
		}

		_, err := d.client.CoreV1().Pods(d.cfg.Namespace).Get(d.ctx, podName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				fmt.Printf("no debug pod found for: %s\n", pod.Name)
				continue
			}
			fmt.Printf("failed to get pod %s: %v\n", podName, err)
			return err
		}

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Pod '%s' with debug container, do you want to clean up", podName),
			IsConfirm: true,
		}

		answer, err := prompt.Run()
		if err != nil {
			if errors.Is(err, promptui.ErrAbort) {
				fmt.Printf("skipping deletion for pod: %s\n", podName)
				continue
			}
			fmt.Printf("error reading input for pod %s: %v\n", podName, err)
			return err
		}

		// Delete pod if user confirms with "y" or "yes"
		if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
			if err := d.client.CoreV1().Pods(d.cfg.Namespace).Delete(d.ctx, podName, metav1.DeleteOptions{}); err != nil {
				fmt.Printf("Failed to delete pod %s: %v\n", podName, err)
			} else {
				fmt.Println("Deleted pod:", podName)
			}
		} else {
			fmt.Println("Skipped pod:", podName)
		}
	}
	return nil
}
