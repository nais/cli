package debug

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pterm/pterm"
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
		pterm.Info.Println("No pods found")
		return nil
	}

	for _, pod := range pods.Items {
		podName := pod.Name
		if d.flags.Copy {
			podName = debuggerContainerName(pod.Name)
		}

		if !d.flags.Copy && len(pod.Spec.EphemeralContainers) == 0 {
			pterm.Info.Printf("No debug container found for: %s\n", pod.Name)
			continue
		}

		_, err := d.client.CoreV1().Pods(d.flags.Namespace).Get(d.ctx, podName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				pterm.Info.Printf("No debug pod found for: %s\n", pod.Name)
				continue
			}
			pterm.Error.Printf("Failed to get pod %s: %v\n", podName, err)
			return err
		}

		confirm, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultText(fmt.Sprintf("Pod '%s' with debug container, do you want to clean up?", podName)).
			Show()

		if !confirm {
			pterm.Info.Printf("Skipping deletion for pod: %s\n", podName)
			continue
		}

		// Delete pod if user confirms
		if err := d.client.CoreV1().Pods(d.flags.Namespace).Delete(d.ctx, podName, metav1.DeleteOptions{}); err != nil {
			pterm.Error.Printf("Failed to delete pod %s: %v\n", podName, err)
		} else {
			pterm.Success.Printf("Deleted pod: %s\n", podName)
		}
	}
	return nil
}
