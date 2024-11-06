package debug

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Debug) Tidy() error {
	pods, err := d.getPods()
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
