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

func getConfig() *rest.Config {
	// setup config file
	var kubeconfig string
	kubeconfig = os.Getenv(KUBECONFIG)
	if kubeconfig == "" {
		log.Fatalf("%s is reqired for debug client to work properly", KUBECONFIG)
	}

	// use the current context in kubeconfig
	configs, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	return configs
}

func SetupClient() ctrl.Client {
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = aiven_nais_io_v1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	config := getConfig()
	client, err := ctrl.New(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}
	return client
}
