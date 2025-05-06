package cli

import "github.com/spf13/cobra"

func validatecmd() *cobra.Command {
	validate := &cobra.Command{
		Use:   "validate file...",
		Short: "Validate nais.yaml configuration",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
			// return validate.Before( ... )
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
			// return validate.Action( ... )
		},
	}
	validate.Flags().String("vars", "", "Path to the `file` containing template variables, must be JSON or YAML format.")
	validate.Flags().StringArray("var", nil, "Template variable in KEY=VALUE form, can be specified multiple times.")
	validate.Flags().String("verbose", "", "Print all the template variables and final resources after templating.")
	return validate
}
