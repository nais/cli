package doctor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/heroku/color"
	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	iconError = "✗"
	iconOK    = "✓"
	iconDot   = "•"
	iconSkip  = "-"
)

type Manager struct {
	log           *logrus.Logger
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	app           *nais_io_v1alpha1.Application
	out           io.Writer
}

func New(log *logrus.Logger, cfg *rest.Config) (*Manager, error) {
	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{
		log:           log,
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,
		out:           os.Stdout,
	}, nil
}

func (m *Manager) Init(ctx context.Context, namespace, appName string) error {
	kapp, err := m.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "nais.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}).Namespace(namespace).Get(ctx, appName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return fmt.Errorf("application %v not found", appName)
		}
		return err
	}

	m.app = &nais_io_v1alpha1.Application{}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(kapp.Object, m.app)
}

func (m *Manager) SetOutput(w io.Writer) {
	m.out = w
}

func (m *Manager) Run(ctx context.Context, verbose bool, skip, only []string) error {
	hasError := false

	fmt.Fprintln(m.out, "Running checks:")
	for _, check := range checks {
		if len(only) > 0 && !contains(only, check.Name()) {
			continue
		}
		if len(skip) > 0 && contains(skip, check.Name()) {
			continue
		}

		cfg := &Config{
			Application:   m.app.DeepCopy(),
			K8sClient:     m.k8sClient,
			DynamicClient: m.dynamicClient,
			Log:           m.log.WithField("check", check.Name()),
			Out:           m.out,
		}
		fmt.Fprint(m.out, "  "+iconDot+" ", color.New(color.Bold).Sprint(check.Name()))
		if verbose {
			fmt.Fprintln(m.out)
		}
		errs := check.Check(ctx, cfg)
		if m.newMethod(verbose, errs) {
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("some checks failed")
	}
	return nil
}

func (m *Manager) newMethod(verbose bool, errs []error) bool {
	handledErrors := 0
	warnings := []error{}
	hasSkip := false
	for _, err := range errs {
		if err == nil {
			continue
		}

		if errors.Is(err, ErrSkip) {
			hasSkip = true
			continue
		} else if errors.Is(err, ErrWarning) {
			warnings = append(warnings, err)
			continue
		}
		if !verbose {
			fmt.Fprintln(m.out, " "+iconError)
		}
		fmt.Fprintln(m.out, color.RedString(err.Error()))
		handledErrors++
	}

	if handledErrors == 0 {
		if !verbose {
			if hasSkip {
				fmt.Fprintln(m.out, " "+color.YellowString(iconSkip))
			} else {
				fmt.Fprintln(m.out, " "+color.GreenString(iconOK))
			}
		}
	}

	for _, err := range warnings {
		fmt.Fprintln(m.out, " "+color.YellowString(err.Error()))
	}
	return handledErrors > 0
}

func List(w io.Writer) {
	checks := checks[:]
	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Name() < checks[j].Name()
	})
	for _, check := range checks {
		fmt.Fprintf(w, "  %v: %v\n", check.Name(), check.Help())
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
