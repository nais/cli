package k8s

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	liberatorscheme "github.com/nais/liberator/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	// Auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var scheme = runtime.NewScheme()

type Client struct {
	ctrl.Client
	CurrentNamespace string
}

func getConfig(overrides []ClientOverride) (*rest.Config, string) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	for _, override := range overrides {
		override(configOverrides)
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatal("Unable to configure Kubernetes client. Check that naisdevice is connected, and your selected context is correct")
	}
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		log.Fatal("Unable to determine current namespace")
	}
	return config, namespace
}

func InitScheme(scheme *runtime.Scheme) {
	_, err := liberatorscheme.AddAll(scheme)
	if err != nil {
		log.Fatalf("error setting up client schema: %s.", err)
	}
}

type ClientOverride func(*clientcmd.ConfigOverrides)

func WithKubeContext(kubeCtx string) ClientOverride {
	return func(overrides *clientcmd.ConfigOverrides) {
		overrides.CurrentContext = kubeCtx
	}
}

func SetupControllerRuntimeClient(overrides ...ClientOverride) *Client {
	ctrllog.SetLogger(logr.FromSlogHandler(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	))

	InitScheme(scheme)
	config, namespace := getConfig(overrides)
	client, err := ctrl.New(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &Client{client, namespace}
}

func SetupClientGo(context string) (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: context,
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get kubeconfig: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("load kubeclient configuration: %w", err)
	}

	return k8sClient, err
}

func GetDefaultContextAndNamespace() (defaultContext string, defaultNamespace string) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	)
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return
	}

	defaultContext = rawConfig.CurrentContext
	if context, exists := rawConfig.Contexts[defaultContext]; exists {
		defaultNamespace = context.Namespace
	}

	return
}
