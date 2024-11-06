package debug

import (
	"errors"
	"fmt"
	"strings"

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

	epHConTotal := 0
	for _, pod := range pods.Items {
		if len(pod.Spec.EphemeralContainers) == 0 {
			continue
		}

		epHConTotal += len(pod.Spec.EphemeralContainers)
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Pod '%s' contains '%d' debug container(s), do you want to clean up", pod.Name, len(pod.Spec.EphemeralContainers)),
			IsConfirm: true,
		}

		answer, err := prompt.Run()
		if err != nil {
			if errors.Is(err, promptui.ErrAbort) {
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
				fmt.Println("Deleted pod:", pod.Name)
			}
		} else {
			fmt.Println("Skipped pod:", pod.Name)
		}
	}

	if epHConTotal == 0 {
		fmt.Printf("Workload '%s' does not contain any debug containers\n", d.cfg.WorkloadName)
	}
	return nil
}
