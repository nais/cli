package doctor

import (
	"fmt"

	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/doctor"
	_ "github.com/nais/cli/pkg/doctor/checks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
)

var doctorCommand = &cobra.Command{
	Use:   "doctor [application] [flags]",
	Short: "Command used to check for potential issues with an application. (BETA)",
	Long: `NAIS Doctor (BETA) will run a series of checks on your application and report any issues it finds.
It will also try to fix any issues it finds, or suggest a fix if it can.`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 1 {
			command.Help()
			fmt.Fprint(command.OutOrStdout(), "\nChecks:\n")
			doctor.List(command.OutOrStdout())
			fmt.Fprintln(command.OutOrStdout())
			return fmt.Errorf("missing argument")
		}
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		verbose := viper.GetBool(cmd.VerboseFlag)
		ctx := command.Context()

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{
			CurrentContext: context,
		}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err := kubeConfig.ClientConfig()
		if err != nil {
			return fmt.Errorf("unable to get kubeconfig: %w", err)
		}

		if namespace == "" {
			namespace, _, err = kubeConfig.Namespace()
			if err != nil {
				return fmt.Errorf("unable to get namespace: %w", err)
			}
		}

		log := logrus.StandardLogger()
		log.SetLevel(logrus.ErrorLevel)
		if verbose {
			log.SetLevel(logrus.DebugLevel)
		}
		doc, err := doctor.New(log, config)
		if err != nil {
			return fmt.Errorf("unable to create doctor: %w", err)
		}

		if err := doc.Init(ctx, namespace, appName); err != nil {
			return fmt.Errorf("unable to init doctor: %w", err)
		}

		if err := doc.Run(ctx, verbose); err != nil {
			return fmt.Errorf("unable to run doctor: %w", err)
		}

		return nil
	},
}

type Config struct{}

func NewConfig() *Config {
	return &Config{}
}

func (c Config) InitCmds(root *cobra.Command) {
	// c.postgres.PersistentFlags().StringP(cmd.NamespaceFlag, "n", "", "Kubernetes namespace where the app is deployed (defaults to the one defined in kubeconfig)")
	// viper.BindPFlag(cmd.NamespaceFlag, c.postgres.PersistentFlags().Lookup(cmd.NamespaceFlag))
	// c.postgres.PersistentFlags().StringP(cmd.ContextFlag, "c", "", "Kubernetes context where the app is deployed (defaults to the one defined in kubeconfig)")
	// viper.BindPFlag(cmd.ContextFlag, c.postgres.PersistentFlags().Lookup(cmd.ContextFlag))

	// c.proxy.Flags().StringP(cmd.PortFlag, "p", "5432", "Local port for the proxy to listen on")
	// viper.BindPFlag(cmd.PortFlag, c.proxy.Flags().Lookup(cmd.PortFlag))
	// c.proxy.Flags().StringP(cmd.HostFlag, "H", "localhost", "Host for the proxy")
	// viper.BindPFlag(cmd.HostFlag, c.proxy.Flags().Lookup(cmd.HostFlag))

	// c.psql.Flags().BoolP(cmd.VerboseFlag, "V", false, "Verbose will also print the proxy logs")
	// viper.BindPFlag(cmd.VerboseFlag, c.psql.Flags().Lookup(cmd.VerboseFlag))

	// c.users.Flags().StringP(cmd.PrivilegeFlag, "", "select", "Privilege level for user in database schema")
	// viper.BindPFlag(cmd.PrivilegeFlag, c.users.Flags().Lookup(cmd.PrivilegeFlag))

	doctorCommand.PersistentFlags().StringP(cmd.NamespaceFlag, "n", "", "Kubernetes namespace where the app is deployed (defaults to the one defined in kubeconfig)")
	viper.BindPFlag(cmd.NamespaceFlag, doctorCommand.PersistentFlags().Lookup(cmd.NamespaceFlag))
	doctorCommand.PersistentFlags().StringP(cmd.ContextFlag, "c", "", "Kubernetes context where the app is deployed (defaults to the one defined in kubeconfig)")
	viper.BindPFlag(cmd.ContextFlag, doctorCommand.PersistentFlags().Lookup(cmd.ContextFlag))
	doctorCommand.PersistentFlags().BoolP(cmd.VerboseFlag, "v", false, "Verbose will also print the debug logs")
	viper.BindPFlag(cmd.VerboseFlag, doctorCommand.PersistentFlags().Lookup(cmd.VerboseFlag))

	root.AddCommand(doctorCommand)
}
