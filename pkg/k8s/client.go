package k8s

import (
	"log"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	liberatorscheme "github.com/nais/liberator/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"
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
}

func getConfig(overrides []ClientOverride) *rest.Config {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	for _, override := range overrides {
		override(configOverrides)
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Unable to configure Kubernetes client. Check that naisdevice is connected, and your selected context is correct")
	}
	return config
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

func SetupClient(overrides ...ClientOverride) ctrl.Client {
	ctrllog.SetLogger(logr.FromSlogHandler(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	))

	InitScheme(scheme)
	config := getConfig(overrides)
	client, err := ctrl.New(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &Client{client}
}
