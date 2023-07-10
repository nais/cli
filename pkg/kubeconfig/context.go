package kubeconfig

import (
	"fmt"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func populateWithContexts(config *clientcmdapi.Config, cluster k8sCluster, email string, options filterOptions) {
	if _, ok := config.Contexts[cluster.Name]; ok && !options.overwrite {
		if options.verbose {
			fmt.Printf("Context %q already exists in kubeconfig, skipping\n", cluster.Name)
		}
		return
	}

	user := email
	if cluster.Kind == kindOnprem {
		user = cluster.User.UserName
	}

	config.Contexts[cluster.Name] = &clientcmdapi.Context{
		Cluster:   cluster.Name,
		AuthInfo:  user,
		Namespace: "default",
	}

	fmt.Printf("Added context %v for %v to config\n", cluster.Name, user)
}
