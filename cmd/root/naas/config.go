package naas

import (
	"fmt"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var naasCommand = &cobra.Command{
	Use:   "naas [command] [args] [flags]",
	Short: "Commands related to a NAAS cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required commands")
	},
}

type Config struct {
	naas          *cobra.Command
	kubeconfigCmd *cobra.Command
}

func NewConfig() *Config {
	return &Config{
		naas:          naasCommand,
		kubeconfigCmd: kubeconfigCmd,
	}
}

func (c Config) InitCmds(root *cobra.Command) {
	c.kubeconfigCmd.Flags().StringP(cmd.TenantFlag, "", "", "Tenant to use")
	_ = c.kubeconfigCmd.Flags().MarkHidden(cmd.TenantFlag)
	viper.BindPFlag(cmd.TenantFlag, c.kubeconfigCmd.Flags().Lookup(cmd.TenantFlag))

	c.kubeconfigCmd.Flags().BoolP(cmd.Force, "", false, "Force overwrite of existing kubeconfig elements")
	viper.BindPFlag(cmd.Force, c.kubeconfigCmd.Flags().Lookup(cmd.Force))

	c.kubeconfigCmd.Flags().StringP(cmd.Email, "", "", "Force kubeconfig to use this email address")
	viper.BindPFlag(cmd.Email, c.kubeconfigCmd.Flags().Lookup(cmd.Email))

	c.kubeconfigCmd.Flags().BoolP(cmd.IncludeManagementFlag, "", false, "Include management clusters")
	viper.BindPFlag(cmd.IncludeManagementFlag, c.kubeconfigCmd.Flags().Lookup(cmd.IncludeManagementFlag))

	root.AddCommand(c.naas)
	c.naas.AddCommand(c.kubeconfigCmd)
}
