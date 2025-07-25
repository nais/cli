package kubeconfig

import (
	"encoding/base64"

	"github.com/nais/naistrix"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func populateWithClusters(config *clientcmdapi.Config, cluster k8sCluster, options filterOptions, out naistrix.Output) error {
	if _, ok := config.Clusters[cluster.Name]; ok && !options.overwrite {
		if options.verbose {
			out.Printf("Cluster %q already exists in kubeconfig, skipping\n", cluster.Name)
		}
		return nil
	}

	var (
		ca  []byte
		err error
	)
	if len(cluster.CA) > 0 {
		ca, err = base64.StdEncoding.DecodeString(cluster.CA)
		if err != nil {
			return err
		}
	}

	kubeconfigCluster := &clientcmdapi.Cluster{
		Server:                   cluster.Endpoint,
		CertificateAuthorityData: ca,
	}

	if cluster.Kind == kindLegacy {
		kubeconfigCluster.CertificateAuthorityData = nil
		kubeconfigCluster.InsecureSkipTLSVerify = true
		kubeconfigCluster.Server = getClusterServerForLegacyGCP(cluster.Name)
	}

	config.Clusters[cluster.Name] = kubeconfigCluster

	out.Printf("Added cluster %v to config\n", cluster.Name)

	return nil
}
