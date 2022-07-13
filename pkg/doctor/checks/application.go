package checks

import (
	"context"
	"fmt"
	"regexp"

	"github.com/nais/cli/pkg/doctor"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
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

func (a *Application) Check(ctx context.Context, cfg *doctor.Config) []error {
	a.cfg = cfg

	errs := []error{
		a.checkConditions(),
		a.checkDeployment(ctx),
		a.checkAnnotations(ctx),
		a.checkIngresses(ctx),
	}

	return errs
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
				a.cfg.Log.Info("Application is reconciling, errors might be temporary if the new deploy was initiated recently.")
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
		if err := a.checkPod(ctx, pod); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) checkPod(ctx context.Context, pod corev1.Pod) error {
	log := a.cfg.Log.WithField("pod", pod.Name)
	log.Debug("checking pod")
	for _, s := range pod.Status.ContainerStatuses {
		if err := a.checkContainer(ctx, log, pod, s); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) checkContainer(ctx context.Context, log *logrus.Entry, pod corev1.Pod, container corev1.ContainerStatus) error {
	log = log.WithField("container", container.Name)
	log.Debug("checking container")
	if container.Ready {
		return nil
	}

	if container.State.Waiting != nil {
		log := log.WithField("reason", container.State.Waiting.Reason)
		log.Info("container is waiting")
		switch container.State.Waiting.Reason {
		case "ContainerCreating":
			log.Info("container is creating")
			return nil
		case "ImagePullBackOff":
			return doctor.ErrorMsg(fmt.Errorf(container.State.Waiting.Message), "container image pull backoff. Ensure that the image exists and are available for the cluster: "+container.Image)
		case "ErrImagePull":
			return doctor.ErrorMsg(fmt.Errorf(container.State.Waiting.Message), "container image pull error: "+container.State.Waiting.Message)
		case "ImageInspectError", "ErrImageNeverPull", "RegistryUnavailable", "InvalidImageName":
			return doctor.ErrorMsg(fmt.Errorf(container.State.Waiting.Message), "container image error: "+container.State.Waiting.Message)
		case "CrashLoopBackOff":
			return a.containerTerminationState(ctx, pod, container, container.LastTerminationState.Terminated)
		default:
			log.Debug("container is waiting for unknown reason")
		}
		log.Info("container is waiting to start: " + container.State.Waiting.Message)
		return nil
	}

	if container.State.Terminated != nil {
		return a.containerTerminationState(ctx, pod, container, container.State.Terminated)
	}

	return nil
}

func (a *Application) checkAnnotations(ctx context.Context) error {
	a.cfg.Log.Debug("checking application annotations")

	checks := map[string]func(string) error{
		"nginx.ingress.kubernetes.io/proxy-body-size": func(s string) error {
			if !regexp.MustCompile(`^\d+[mM]$`).MatchString(s) {
				err := fmt.Errorf("invalid value for nginx.ingress.kubernetes.io/proxy-body-size: %s. Must be digits followed by a lowercase `m`", s)
				return doctor.ErrorMsg(err, err.Error())
			}
			return nil
		},
	}

	if a.cfg.Application.Annotations == nil {
		a.cfg.Log.Info("no annotations on application")
		return nil
	}

	for k, v := range checks {
		if s, ok := a.cfg.Application.Annotations[k]; ok {
			if err := v(s); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Application) containerTerminationState(ctx context.Context, pod corev1.Pod, container corev1.ContainerStatus, state *corev1.ContainerStateTerminated) error {
	if state == nil {
		return nil
	}

	if state.Reason != "Error" {
		return nil
	}

	fmt.Fprintln(a.cfg.Out, "\nContainer", container.Name, "has termintaed with exit code", state.ExitCode, ". Changes likely required.\nLast 50 lines of logs:")
	res := a.cfg.K8sClient.CoreV1().Pods(a.cfg.Application.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: container.Name,
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

func (a *Application) checkIngresses(ctx context.Context) error {
	a.cfg.Log.Debug("checking ingresses")

	for _, ing := range a.cfg.Application.Spec.Ingresses {
		if err := a.checkIngress(ctx, ing); err != nil {
			return err
		}
	}

	return nil
}

var regDeprecatedIngresses = regexp.MustCompile(`\.(?:dev|prod)\-gcp\.nais\.io(\/|$)`)

func (a *Application) checkIngress(ctx context.Context, ing nais_io_v1.Ingress) error {
	log := a.cfg.Log.WithField("ingress", ing)
	log.Debug("checking ingress")

	if regDeprecatedIngresses.MatchString(string(ing)) {
		return doctor.ErrorMsg(doctor.ErrWarning, fmt.Sprintf("deprecated ingress %q. Please update it.", ing))
	}

	return nil
}
