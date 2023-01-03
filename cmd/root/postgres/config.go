package postgres

import (
	"fmt"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var postgresCommand = &cobra.Command{
	Use:   "postgres [command] [args] [flags]",
	Short: "Command used for connecting to Postgres",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required commands")
	},
}

var usersCommand = &cobra.Command{
	Use:   "users [command] [args] [flags]",
	Short: "Command used for listing and adding users to database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing required commands")
	},
}

type Config struct {
	postgres  *cobra.Command
	proxy     *cobra.Command
	grant     *cobra.Command
	prepare   *cobra.Command
	psql      *cobra.Command
	users     *cobra.Command
	listUsers *cobra.Command
	addUser   *cobra.Command
}

func NewConfig() *Config {
	return &Config{
		postgres:  postgresCommand,
		proxy:     proxyCmd,
		grant:     grantCmd,
		prepare:   prepareCmd,
		psql:      psqlCmd,
		users:     usersCommand,
		listUsers: listUsersCmd,
		addUser:   addUserCmd,
	}
}

func (c Config) InitCmds(root *cobra.Command) {
	c.postgres.PersistentFlags().StringP(cmd.NamespaceFlag, "n", "", "Kubernetes namespace where the app is deployed (defaults to the one defined in kubeconfig)")
	viper.BindPFlag(cmd.NamespaceFlag, c.postgres.PersistentFlags().Lookup(cmd.NamespaceFlag))
	c.postgres.PersistentFlags().StringP(cmd.ContextFlag, "c", "", "Kubernetes context where the app is deployed (defaults to the one defined in kubeconfig)")
	viper.BindPFlag(cmd.ContextFlag, c.postgres.PersistentFlags().Lookup(cmd.ContextFlag))

	c.proxy.Flags().StringP(cmd.PortFlag, "p", "5432", "Local port for the proxy to listen on")
	viper.BindPFlag(cmd.PortFlag, c.proxy.Flags().Lookup(cmd.PortFlag))
	c.proxy.Flags().StringP(cmd.HostFlag, "H", "localhost", "Host for the proxy")
	viper.BindPFlag(cmd.HostFlag, c.proxy.Flags().Lookup(cmd.HostFlag))

	c.psql.Flags().BoolP(cmd.VerboseFlag, "V", false, "Verbose will also print the proxy logs")
	viper.BindPFlag(cmd.VerboseFlag, c.psql.Flags().Lookup(cmd.VerboseFlag))

	c.users.Flags().StringP(cmd.PrivilegeFlag, "", "select", "Privilege level for user in database schema")
	viper.BindPFlag(cmd.PrivilegeFlag, c.users.Flags().Lookup(cmd.PrivilegeFlag))

	c.prepare.Flags().BoolP(cmd.AllPrivs, "", false, "Should all privileges be given?")
	viper.BindPFlag(cmd.AllPrivs, c.prepare.Flags().Lookup(cmd.AllPrivs))

	root.AddCommand(c.postgres)
	c.postgres.AddCommand(c.proxy)
	c.postgres.AddCommand(c.grant)
	c.postgres.AddCommand(c.prepare)
	c.postgres.AddCommand(c.psql)
	c.postgres.AddCommand(c.users)
	c.users.AddCommand(c.listUsers)
	c.users.AddCommand(c.addUser)
}
