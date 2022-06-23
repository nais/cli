package aiven

import (
	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	aiven  *cobra.Command
	create *cobra.Command
	get    *cobra.Command
	tidy   *cobra.Command
}

func NewConfig() *Config {
	return &Config{
		aiven:  aivenCommand,
		create: createCmd,
		get:    getCmd,
		tidy:   tidyCmd,
	}
}

func (a Config) InitCmds(root *cobra.Command) {
	a.create.Flags().StringP(cmd.PoolFlag, "p", "nav-dev", "Preferred kafka pool to connect to (optional)")
	viper.BindPFlag(cmd.PoolFlag, a.create.Flags().Lookup(cmd.PoolFlag))

	a.create.Flags().IntP(cmd.ExpireFlag, "e", 1, "Time in days the created secret should be valid (optional)")
	viper.BindPFlag(cmd.ExpireFlag, a.create.Flags().Lookup(cmd.ExpireFlag))

	a.create.Flags().StringP(cmd.SecretNameFlag, "s", "", "Preferred secret-name instead of generated (optional)")
	viper.BindPFlag(cmd.SecretNameFlag, a.create.Flags().Lookup(cmd.SecretNameFlag))

	a.create.Flags().StringP(cmd.InstanceFlag, "i", "", "Instance to connect to (required for OpenSearch)")
	viper.BindPFlag(cmd.InstanceFlag, a.create.Flags().Lookup(cmd.InstanceFlag))

	a.create.Flags().StringP(cmd.AccessFlag, "a", "read", "Type of access needed. Supported values: read, write, readwrite, admin (optional)")
	viper.BindPFlag(cmd.AccessFlag, a.create.Flags().Lookup(cmd.AccessFlag))

	root.AddCommand(a.aiven)
	a.aiven.AddCommand(a.create)
	a.aiven.AddCommand(a.get)
	a.aiven.AddCommand(a.tidy)
}
