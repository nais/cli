package naas

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/nais/cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type project struct {
	ID     string
	Tenant string
	Kind   string
	Name   string
}

type userInfo struct {
	ServerID string `json:"serverID"`
	ClientID string `json:"clientID"`
	TenantID string `json:"tenantID"`
	UserName string `json:"userName"`
}

type clusterEntry struct {
	Name     string
	Endpoint string
	Location string
	CA       string

	// only used for on-prem clusters
	User *userInfo
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
