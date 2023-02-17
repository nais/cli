package naas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

type kubeConfigSync struct {
	prefix            bool
	tenant            string
	email             string
	force             bool
	includeManagement bool

	log logrus.FieldLogger
}

func (k *kubeConfigSync) Run(ctx context.Context) error {
	projects, err := k.projects(ctx)
	if err != nil {
		return err
	}

	configLoad := clientcmd.NewDefaultClientConfigLoadingRules()

	// If KUBECONFIG is set, but the file does not exist, clientcmd will throw a warning.
	// since we're creating the file, we can safely ignore this warning.
	klog.SetLogger(logr.Discard())

	config, err := configLoad.Load()
	if err != nil {
		return err
	}

	clusters := []clusterEntry{}
	for _, project := range projects {
		cluster, err := k.clusters(ctx, project)
		if err != nil {
			return err
		}

		onprem, err := k.onpremClusters(ctx, project)
		if err != nil {
			return err
		}

		clusters = append(clusters, cluster...)
		if len(onprem) > 0 {
			clusters = append(clusters, onprem...)
		}
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters found")
	}

	if err := k.addUsersToConfig(config, clusters); err != nil {
		return err
	}

	if err := k.addToConfig(config, clusters); err != nil {
		return err
	}

	if err := clientcmd.WriteToFile(*config, configLoad.GetDefaultFilename()); err != nil {
		return err
	}

	fmt.Println("kubeconfig written to", configLoad.GetDefaultFilename())

	for _, user := range config.AuthInfos {
		if _, err := exec.LookPath(user.Exec.Command); err != nil {
			fmt.Fprintf(os.Stderr, "\nWARNING: %v not found in PATH.\n", user.Exec.Command)
			fmt.Fprintln(os.Stderr, user.Exec.InstallHint)
		}
	}

	return nil
}

func (k *kubeConfigSync) projects(ctx context.Context) ([]project, error) {
	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	projects := []project{}
	filter := "(labels.naiscluster:true OR labels.kind:*)"
	if !k.includeManagement {
		filter += " labels.environment:*"
	}
	if k.tenant != "" {
		filter += " labels.tenant:" + k.tenant
	}
	call := svc.Projects.List().Filter(filter)
	for {
		ps, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, p := range ps.Projects {
			projects = append(projects, project{
				ID:     p.ProjectId,
				Tenant: p.Labels["tenant"],
				Kind:   p.Labels["kind"],
				Name:   p.Labels["environment"],
			})
		}
		if ps.NextPageToken == "" {
			break
		}
		call.PageToken(ps.NextPageToken)
	}

	return projects, nil
}

func (k *kubeConfigSync) clusters(ctx context.Context, project project) ([]clusterEntry, error) {
	svc, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	call := svc.Projects.Locations.Clusters.List("projects/" + project.ID + "/locations/-")
	clusters, err := call.Do()
	if err != nil {
		return nil, err
	}

	ret := []clusterEntry{}
	for _, cluster := range clusters.Clusters {
		name := cluster.Name
		if k.prefix {
			name = project.Tenant + "-" + strings.TrimPrefix(name, "nais-")
		}
		ret = append(ret, clusterEntry{
			Name:     name,
			Endpoint: "https://" + cluster.Endpoint,
			Location: cluster.Location,
			CA:       cluster.MasterAuth.ClusterCaCertificate,
		})
	}
	return ret, nil
}

func (k *kubeConfigSync) onpremClusters(ctx context.Context, project project) ([]clusterEntry, error) {
	if project.Kind != "onprem" {
		return nil, nil
	}

	svc, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}
	proj, err := svc.Projects.Get(project.ID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	ret := []clusterEntry{}
	for _, meta := range proj.CommonInstanceMetadata.Items {
		if meta.Key != "kubeconfig" || meta.Value == nil {
			continue
		}

		config := &struct {
			ServerID string `json:"serverID"`
			ClientID string `json:"clientID"`
			TenantID string `json:"tenantID"`
			URL      string `json:"url"`
			UserName string `json:"userName"`
		}{}
		if err := json.Unmarshal([]byte(*meta.Value), &config); err != nil {
			return nil, err
		}

		name := project.Name
		if k.prefix {
			name = project.Tenant + "-" + strings.TrimPrefix(name, "nais-")
		}
		ret = append(ret, clusterEntry{
			Name:     name,
			Endpoint: config.URL,
			User: &userInfo{
				ServerID: config.ServerID,
				ClientID: config.ClientID,
				TenantID: config.TenantID,
				UserName: config.UserName,
			},
		})

		return ret, nil

	}

	return ret, nil
}

func (k *kubeConfigSync) addUsersToConfig(config *clientcmdapi.Config, clusters []clusterEntry) error {
	hasGCP := false
	onpremUsers := map[string]*userInfo{}
	for _, cluster := range clusters {
		if cluster.User == nil {
			hasGCP = true
		} else {
			onpremUsers[cluster.User.UserName] = cluster.User
		}
	}

	if hasGCP {
		if _, ok := config.AuthInfos[k.email]; ok && !k.force {
			k.log.Info("user %q already exists in kubeconfig, skipping", k.email)
		} else {
			config.AuthInfos[k.email] = &clientcmdapi.AuthInfo{
				Exec: &clientcmdapi.ExecConfig{
					APIVersion: "client.authentication.k8s.io/v1beta1",
					Args:       nil,
					Command:    "gke-gcloud-auth-plugin",
					Env: []clientcmdapi.ExecEnvVar{
						{
							Name:  "CLOUDSDK_CORE_ACCOUNT",
							Value: k.email,
						},
					},
					InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following\nhttps://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke",
					InteractiveMode:    clientcmdapi.IfAvailableExecInteractiveMode,
					ProvideClusterInfo: true,
				},
			}
		}
	}

	keys := make([]string, 0, len(onpremUsers))
	for k := range onpremUsers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		user := onpremUsers[key]
		if _, ok := config.AuthInfos[user.UserName]; ok && !k.force {
			k.log.Infof("user %q already exists in kubeconfig, skipping", user.UserName)
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
	}

	return nil
}

func (k *kubeConfigSync) addToConfig(config *clientcmdapi.Config, clusters []clusterEntry) error {
	for _, cluster := range clusters {
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

		if _, ok := config.Clusters[cluster.Name]; ok && !k.force {
			k.log.Infof("cluster %q already exists in kubeconfig, skipping", cluster.Name)
		} else {
			config.Clusters[cluster.Name] = &clientcmdapi.Cluster{
				Server:                   cluster.Endpoint,
				CertificateAuthorityData: ca,
			}
		}

		user := k.email
		if cluster.User != nil {
			user = cluster.User.UserName
		}

		if _, ok := config.Contexts[cluster.Name]; ok && !k.force {
			k.log.Infof("context %q already exists in kubeconfig, skipping", cluster.Name)
		} else {
			config.Contexts[cluster.Name] = &clientcmdapi.Context{
				Cluster:   cluster.Name,
				AuthInfo:  user,
				Namespace: "default",
			}
		}
	}

	if config.CurrentContext == "" && len(clusters) > 0 {
		config.CurrentContext = clusters[0].Name
	}

	return nil
}
