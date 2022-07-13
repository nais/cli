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
		skip := viper.GetStringSlice("skip")
		only := viper.GetStringSlice("only")
		ctx := command.Context()

		if len(skip) > 0 && len(only) > 0 {
			return fmt.Errorf("--skip and --only can not be used together")
		}

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

		if err := doc.Run(ctx, verbose, skip, only); err != nil {
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
	doctorCommand.PersistentFlags().StringSlice("only", nil, "Only run the listed checks")
	viper.BindPFlag("only", doctorCommand.PersistentFlags().Lookup("only"))
	doctorCommand.PersistentFlags().StringSlice("skip", nil, "skip running the listed checks")
	viper.BindPFlag("skip", doctorCommand.PersistentFlags().Lookup("skip"))

	doctorCommand.PersistentFlags().StringP(cmd.NamespaceFlag, "n", "", "Kubernetes namespace where the app is deployed (defaults to the one defined in kubeconfig)")
	viper.BindPFlag(cmd.NamespaceFlag, doctorCommand.PersistentFlags().Lookup(cmd.NamespaceFlag))
	doctorCommand.PersistentFlags().StringP(cmd.ContextFlag, "c", "", "Kubernetes context where the app is deployed (defaults to the one defined in kubeconfig)")
	viper.BindPFlag(cmd.ContextFlag, doctorCommand.PersistentFlags().Lookup(cmd.ContextFlag))
	doctorCommand.PersistentFlags().BoolP(cmd.VerboseFlag, "v", false, "Verbose will also print the debug logs")
	viper.BindPFlag(cmd.VerboseFlag, doctorCommand.PersistentFlags().Lookup(cmd.VerboseFlag))

	root.AddCommand(doctorCommand)
}
