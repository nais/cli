package client

import (
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

var scheme = runtime.NewScheme()

const (
	KUBECONFIG = "KUBECONFIG"
)

type AivenClient struct {
	ctrl.Client
}

func getConfig() *rest.Config {
	// setup config file
	var kubeconfig string
	kubeconfig = os.Getenv(KUBECONFIG)
	if kubeconfig == "" {
		log.Fatalf("%s environment variable is reqired for client to work properly.\n"+
			"naisdevice: 1. Installed? 2. Running? 3. Connected?", KUBECONFIG)
	}

	// use the current context in kubeconfig
	configs, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("authentication: update your kubectcl kubeconfig: %s", err)
	}
	return configs
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
