package kubeconfig

import (
	"fmt"

	"github.com/nais/cli/pkg/gcp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func addUsers(config *clientcmdapi.Config, clusters []gcp.Cluster, email string, overwrite, includeOnprem, verbose bool) error {
	addGCPUser(config, email, overwrite, verbose)

	if !includeOnprem {
		return nil
	}

	return addOnpremUser(config, clusters, overwrite, verbose)
}

func addOnpremUser(config *clientcmdapi.Config, clusters []gcp.Cluster, overwrite, verbose bool) error {
	for _, cluster := range clusters {
		if cluster.Kind == gcp.KindOnprem {
			user := cluster.User
			if _, ok := config.AuthInfos[user.UserName]; ok && !overwrite {
				if verbose {
					fmt.Printf("User %q already exists in kubeconfig, skipping\n", user.UserName)
				}
				continue
			}

			config.AuthInfos[user.UserName] = &clientcmdapi.AuthInfo{
				Exec: &clientcmdapi.ExecConfig{
					APIVersion: "client.authentication.k8s.io/v1beta1",
					Args: []string{
						"get-token",
						"--login",
						"devicecode",
						"--server-id",
						user.ServerID,
						"--client-id",
						user.ClientID,
						"--tenant-id",
						user.TenantID,
						"--legacy",
					},
					Command:            "kubelogin",
					InstallHint:        "Install kubelogin for use with kubectl by following\nhttps://github.com/Azure/kubelogin#getting-started",
					InteractiveMode:    clientcmdapi.IfAvailableExecInteractiveMode,
					ProvideClusterInfo: false,
				},
			}

			fmt.Printf("Added user %v to config\n", user.UserName)

			return nil
		}
	}
	return nil
}

func addGCPUser(config *clientcmdapi.Config, email string, overwrite, verbose bool) {
	if _, ok := config.AuthInfos[email]; ok && !overwrite {
		if verbose {
			fmt.Printf("User %q already exists in kubeconfig, skipping\n", email)
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

	fmt.Printf("Added user %v to config\n", email)
}
