package checks

import (
	"context"
	"fmt"

	"github.com/nais/cli/pkg/doctor"
	appsv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type Application struct {
	cfg *doctor.Config
}

func init() {
	doctor.AddCheck(&Application{})
}

func (a *Application) Name() string {
	return "application"
}

func (a *Application) Help() string {
	return "Check common issues that prevent the application from functioning."
}

func (a *Application) Check(ctx context.Context, cfg *doctor.Config) error {
	a.cfg = cfg

	if err := a.checkConditions(); err != nil {
		return err
	}

	if err := a.checkDeployment(ctx); err != nil {
		return err
	}

	return nil
}

func (a *Application) checkConditions() error {
	a.cfg.Log.Debug("checking application conditions")
	if a.cfg.Application.Status.Conditions == nil {
		return doctor.ErrorMsg(fmt.Errorf("application has no condition"), "application has no condition, this might indicate that the application hasn't completed initial deploy.")
	}
	for _, con := range *a.cfg.Application.Status.Conditions {
		switch con.Type {
		case "Reconciling":
			if con.Status == "True" {
				return doctor.ErrorMsg(fmt.Errorf("application is reconciling"), "application is reconciling, wait for it to complete.")
			}
		}
	}
	return nil
}

func (a *Application) checkDeployment(ctx context.Context) error {
	a.cfg.Log.Debug("checking deployment")
	depl, err := a.cfg.K8sClient.AppsV1().Deployments(a.cfg.Application.Namespace).Get(ctx, a.cfg.Application.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return doctor.ErrorMsg(err, "application deployment not found. Try redeploying the application.")
		}
		return err
	}

	// Check deployment for certain conditions
	_ = depl

	// find pods
	selector := "app=" + a.cfg.Application.Name
	a.cfg.Log.WithField("selector", selector).Debug("finding pods")
	pods, err := a.cfg.K8sClient.CoreV1().Pods(a.cfg.Application.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return doctor.ErrorMsg(err, "application pods not found. Manual intervention likely required.")
		}
		return err
	}

	for _, pod := range pods.Items {
		a.checkPod(ctx, pod)
	}

	return nil
}

func (a *Application) checkPod(ctx context.Context, pod appsv1.Pod) error {
	log := a.cfg.Log.WithField("pod", pod.Name)
	log.Debug("checking pod")
	for _, s := range pod.Status.ContainerStatuses {
		if s.Name == a.cfg.Application.Name {
			log.Debug("found container matching app name")

			if !s.Ready {
				if s.State.Terminated == nil {
					continue
				}

				if s.State.Terminated.Reason == "Completed" {
					continue
				}

				if s.State.Terminated.Reason == "Error" {
					fmt.Fprintln(a.cfg.Out, "\nApplication has termintaed with exit code", s.State.Terminated.ExitCode, ". Changes likely required.\nLast 50 lines of logs:")
					res := a.cfg.K8sClient.CoreV1().Pods(a.cfg.Application.Namespace).GetLogs(pod.Name, &appsv1.PodLogOptions{
						Container: s.Name,
						TailLines: pointer.Int64Ptr(50),
					}).Do(ctx)

					if err := res.Error(); err != nil {
						fmt.Fprintln(a.cfg.Out, "Error getting logs:", err)
					} else {
						logs, err := res.Raw()
						if err != nil {
							fmt.Fprintln(a.cfg.Out, "Error getting logs:", err)
						} else {
							fmt.Fprintln(a.cfg.Out, string(logs))
						}
					}
					return doctor.ErrorMsg(fmt.Errorf("pod %s has error state", pod.Name), "pod has error state, likely exited with an error.")
				}
			}
		}
	}
	return nil
}
