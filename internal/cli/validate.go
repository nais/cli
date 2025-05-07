package cli

import (
	validatecmd "github.com/nais/cli/internal/validate"
	"github.com/spf13/cobra"
)

func validate() *cobra.Command {
	flags := validatecmd.Flags{}
	cmd := &cobra.Command{
		Use:   "validate file...",
		Short: "Validate one or more Nais manifest files",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
			var comps []cobra.Completion
			if len(args) == 0 {
				comps = cobra.AppendActiveHelp(comps, "Choose at least one manifest file to validate")
			}
			return comps, cobra.ShellCompDirectiveDefault
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			flags.Verbose = verbose
			return validatecmd.Run(args, flags)
		},
	}

	fs := cmd.Flags()
	fs.StringVarP(&flags.VarsFilePath, "vars", "f", "", "Path to the `file` containing template variables, must be JSON or YAML format.")
	fs.StringSliceVar(&flags.Vars, "var", nil, "Template variable in `KEY=VALUE` form. This flag can be repeated.")

	return cmd
}
