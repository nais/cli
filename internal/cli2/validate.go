package cli2

import "github.com/spf13/cobra"

func validatecmd() *cobra.Command {
	validate := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate nais.yaml configuration",
		// Before: validate.Before,
		// Run:    validate.Action,
	}
	validate.Flags().String("vars", "", "Path to the `file` containing template variables, must be JSON or YAML format.")
	validate.Flags().StringArray("var", nil, "Template variable in KEY=VALUE form, can be specified multiple times.")
	validate.Flags().String("verbose", "", "Print all the template variables and final resources after templating.")
	return validate
}
