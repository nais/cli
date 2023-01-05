package naas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/nais/cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/cloudresourcemanager/v1"
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

type project struct {
	ID     string
	Tenant string
}

type clusterEntry struct {
	Name     string
	Endpoint string
	Location string
	CA       string
}

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig [flags]",
	Short: "Create a kubeconfig file for connecting to available clusters.",
	Long: `Create a kubeconfig file for connecting to available clusters.
	This requires that you have the gcloud command line tool installed, configured and logged
	in using:
	gcloud auth login --update-adc`,
	Args: cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		ctx := command.Context()

		email := viper.GetString("email")
		if email == "" {
			var err error
			email, err = gcloudEmail(ctx)
			if err != nil {
				return err
			}
		}

		log := logrus.New()
		log.Level = logrus.WarnLevel

		kcs := &kubeConfigSync{
			tenant:            viper.GetString(cmd.TenantFlag),
			log:               log,
			email:             email,
			force:             viper.GetBool("force"),
			includeManagement: viper.GetBool(cmd.IncludeManagementFlag),
		}

		kcs.prefix = kcs.tenant != ""

		return kcs.Run(ctx)
	},
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

		clusters = append(clusters, cluster...)
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters found")
	}

	if err := k.addToConfig(config, clusters); err != nil {
		return err
	}

	return clientcmd.WriteToFile(*config, configLoad.GetDefaultFilename())
}

func (k *kubeConfigSync) projects(ctx context.Context) ([]project, error) {
	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	projects := []project{}
	filter := "labels.naiscluster:true"
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

func (k *kubeConfigSync) addToConfig(config *clientcmdapi.Config, clusters []clusterEntry) error {
	if _, ok := config.AuthInfos[k.email]; ok && !k.force {
		k.log.Warnf("user %q already exists in kubeconfig, skipping", k.email)
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

	for _, cluster := range clusters {
		ca, err := base64.StdEncoding.DecodeString(cluster.CA)
		if err != nil {
			return err
		}

		if _, ok := config.Clusters[cluster.Name]; ok && !k.force {
			k.log.Warnf("cluster %q already exists in kubeconfig, skipping", cluster.Name)
		} else {
			config.Clusters[cluster.Name] = &clientcmdapi.Cluster{
				Server:                   cluster.Endpoint,
				CertificateAuthorityData: ca,
			}
		}

		if _, ok := config.Contexts[cluster.Name]; ok && !k.force {
			k.log.Warnf("context %q already exists in kubeconfig, skipping", cluster.Name)
		} else {
			config.Contexts[cluster.Name] = &clientcmdapi.Context{
				Cluster:   cluster.Name,
				AuthInfo:  k.email,
				Namespace: "default",
			}
		}
	}

	if config.CurrentContext == "" && len(clusters) > 0 {
		config.CurrentContext = clusters[0].Name
	}

	return nil
}

func gcloudEmail(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "config-helper", "--format", "json")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return "", fmt.Errorf("gcloud command failed: %s", string(b))
	}

	cfg := &struct {
		Configuration struct {
			Properties struct {
				Core struct {
					Account string `json:"account"`
				} `json:"core"`
			} `json:"properties"`
		} `json:"configuration"`
	}{}

	if err := json.Unmarshal(b, cfg); err != nil {
		return "", err
	}
	return cfg.Configuration.Properties.Core.Account, nil
}
