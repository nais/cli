package validate

import (
	"fmt"

	"github.com/nais/cli/pkg/validate"
	"github.com/spf13/cobra"
)

func InitValidateCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(validateCommand)
}

var validateCommand = &cobra.Command{
	Use:   "validate [config...]",
	Short: "Validate nais.yaml configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no config files provided")
		}

		return validate.NaisConfig(args)
	},
}
