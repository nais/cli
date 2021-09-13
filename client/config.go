package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

const (
	KUBECONFIG = "KUBECONFIG"
)

func GetConfig() *rest.Config {
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

func StandardClient() *kubernetes.Clientset {
	config := GetConfig()
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil
	}
	return client

}
