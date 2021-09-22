package client

import (
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// Auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var scheme = runtime.NewScheme()

type AivenClient struct {
	ctrl.Client
}

func getConfig() *rest.Config {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Unable to configure Kubernetes client. Check that naisdevice is connected, and your selected context is correct")
	}
	return config
}

func InitScheme(scheme *runtime.Scheme) {
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("error setting up client schema: %s.", err)
	}

	err = aiven_nais_io_v1.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("error setting up aiven application schema: %s.", err)
	}
}

func SetupClient() ctrl.Client {
	InitScheme(scheme)
	config := getConfig()
	client, err := ctrl.New(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &AivenClient{client}
}
