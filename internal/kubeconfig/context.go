package kubeconfig

import (
	"github.com/nais/cli/pkg/cli"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func populateWithContexts(config *clientcmdapi.Config, cluster k8sCluster, email string, options filterOptions, out cli.Output) {
	if _, ok := config.Contexts[cluster.Name]; ok && !options.overwrite {
		if options.verbose {
			out.Printf("Context %q already exists in kubeconfig, skipping\n", cluster.Name)
		}
		return
	}

	user := email
	if cluster.Kind == kindOnprem {
		user = cluster.User.UserName
	}

	context := &clientcmdapi.Context{
		Cluster:   cluster.Name,
		AuthInfo:  user,
		Namespace: "default",
	}
	if existingCtx, ok := config.Contexts[cluster.Name]; ok && existingCtx.Namespace != "" {
		if options.verbose {
			out.Printf("Preserving namespace %q for existing context %q\n", existingCtx.Namespace, cluster.Name)
		}
		context.Namespace = existingCtx.Namespace
	}

	config.Contexts[cluster.Name] = context

	out.Printf("Added context %v for %v to config\n", cluster.Name, user)
}
