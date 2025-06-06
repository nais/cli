package kubeconfig

import (
	"github.com/nais/cli/internal/cli"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func addUsers(config *clientcmdapi.Config, clusters []k8sCluster, email string, options filterOptions, out cli.Output) error {
	addGCPUser(config, email, options, out)

	if !options.includeOnprem {
		return nil
	}

	return addOnpremUser(config, clusters, options, out)
}

func addOnpremUser(config *clientcmdapi.Config, clusters []k8sCluster, options filterOptions, out cli.Output) error {
	for _, cluster := range clusters {
		if cluster.Kind == kindOnprem {
			user := cluster.User
			if _, ok := config.AuthInfos[user.UserName]; ok && !options.overwrite {
				if options.verbose {
					out.Printf("User %q already exists in kubeconfig, skipping\n", user.UserName)
				}
				continue
			}

			config.AuthInfos[user.UserName] = &clientcmdapi.AuthInfo{
				Exec: &clientcmdapi.ExecConfig{
					APIVersion: "client.authentication.k8s.io/v1beta1",
					Args: []string{
						"get-token",
						"--login",
						"interactive",
						"--server-id",
						user.ServerID,
						"--client-id",
						user.ClientID,
						"--tenant-id",
						user.TenantID,
					},
					Command:            "kubelogin",
					InstallHint:        "Install kubelogin for use with kubectl by following\nhttps://github.com/Azure/kubelogin#getting-started",
					InteractiveMode:    clientcmdapi.IfAvailableExecInteractiveMode,
					ProvideClusterInfo: false,
				},
			}

			out.Printf("Added user %v to config\n", user.UserName)

			return nil
		}
	}
	return nil
}

func addGCPUser(config *clientcmdapi.Config, email string, options filterOptions, out cli.Output) {
	if _, ok := config.AuthInfos[email]; ok && !options.overwrite {
		if options.verbose {
			out.Printf("User %q already exists in kubeconfig, skipping\n", email)
		}
		return
	}

	config.AuthInfos[email] = &clientcmdapi.AuthInfo{
		Exec: &clientcmdapi.ExecConfig{
			APIVersion: "client.authentication.k8s.io/v1beta1",
			Args:       nil,
			Command:    "gke-gcloud-auth-plugin",
			Env: []clientcmdapi.ExecEnvVar{
				{
					Name:  "CLOUDSDK_CORE_ACCOUNT",
					Value: email,
				},
			},
			InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following\nhttps://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke",
			InteractiveMode:    clientcmdapi.IfAvailableExecInteractiveMode,
			ProvideClusterInfo: true,
		},
	}

	out.Printf("Added user %v to config\n", email)
}
