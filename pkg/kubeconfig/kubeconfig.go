package kubeconfig

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/go-logr/logr"
	kubeClient "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

func CreateKubeconfig(ctx context.Context, email string, opts ...FilterOption) error {
	var options filterOptions
	for _, opt := range DefaultFilterOptions {
		opt(&options)
	}
	for _, opt := range opts {
		opt(&options)
	}

	configLoad := kubeClient.NewDefaultClientConfigLoadingRules()

	// If KUBECONFIG is set, but the file does not exist, kubeClient will throw a warning.
	// since we're creating the file, we can safely ignore this warning.
	klog.SetLogger(logr.Discard())

	config, err := configLoad.Load()
	if err != nil {
		return err
	}

	if options.fromScratch {
		config.AuthInfos = map[string]*api.AuthInfo{}
		config.Contexts = map[string]*api.Context{}
		config.Clusters = map[string]*api.Cluster{}
	}

	fmt.Println("Retreiving clusters")
	clusters, err := getClustersFromGCP(ctx, options)
	if err != nil {
		return err
	}
	fmt.Printf("Found %v clusters\n", len(clusters))

	err = addUsers(config, clusters, email, options)
	if err != nil {
		return err
	}

	err = populateKubeconfig(config, clusters, email, options)
	if err != nil {
		return err
	}

	err = kubeClient.WriteToFile(*config, configLoad.GetDefaultFilename())
	if err != nil {
		return err
	}

	fmt.Println("Kubeconfig written to", configLoad.GetDefaultFilename())

	for _, user := range config.AuthInfos {
		if user == nil || user.Exec == nil {
			continue
		}
		_, err = exec.LookPath(user.Exec.Command)
		if err != nil {
			fmt.Printf("%v\nWARNING: %v not found in PATH.\n", os.Stderr, user.Exec.Command)
			fmt.Printf("%v\n%v\n", os.Stderr, user.Exec.InstallHint)
		}
	}
	return nil
}

func populateKubeconfig(config *api.Config, clusters []k8sCluster, email string, options filterOptions) error {
	for _, cluster := range clusters {
		if slices.Contains(options.excludeClusters, cluster.Name) {
			if options.verbose {
				fmt.Printf("Cluster %q is excluded, skipping\n", cluster.Name)
			}
			continue
		}

		err := populateWithClusters(config, cluster, options)
		if err != nil {
			return err
		}

		populateWithContexts(config, cluster, email, options)
	}

	return nil
}
