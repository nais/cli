package cli

import "github.com/spf13/cobra"

func validate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate file...",
		Short: "Validate nais.yaml configuration",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().String("vars", "", "Path to the `file` containing template variables, must be JSON or YAML format.")
	cmd.Flags().StringArray("var", nil, "Template variable in KEY=VALUE form, can be specified multiple times.")
	cmd.Flags().String("verbose", "", "Print all the template variables and final resources after templating.")

	return cmd
}
