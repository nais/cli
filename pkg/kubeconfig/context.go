package kubeconfig

import (
	"fmt"

	"github.com/nais/cli/pkg/gcp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func addContext(config *clientcmdapi.Config, cluster gcp.Cluster, overwrite, verbose bool, email string) {
	if _, ok := config.Contexts[cluster.Name]; ok && !overwrite {
		if verbose {
			fmt.Printf("Context %q already exists in kubeconfig, skipping\n", cluster.Name)
		}
		return
	}

	user := email
	if cluster.Kind == gcp.KindOnprem {
		user = cluster.User.UserName
	}

	config.Contexts[cluster.Name] = &clientcmdapi.Context{
		Cluster:   cluster.Name,
		AuthInfo:  user,
		Namespace: "default",
	}

	fmt.Printf("Added context %v for %v to config\n", cluster.Name, user)
}
