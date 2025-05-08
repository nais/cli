package cli

import (
	"github.com/nais/cli/internal/root"
	validatecmd "github.com/nais/cli/internal/validate"
	"github.com/spf13/cobra"
)

func validate(rootFlags root.Flags) *cobra.Command {
	cmdFlags := validatecmd.Flags{Flags: rootFlags}
	cmd := &cobra.Command{
		Use:   "validate FILE...",
		Short: "Validate one or more Nais manifest files.",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(*cobra.Command, []string, string) ([]cobra.Completion, cobra.ShellCompDirective) {
			comps := []cobra.Completion{"yaml", "yml", "json"}
			comps = cobra.AppendActiveHelp(comps, "Choose one or more manifest files to validate (*.yaml, *.yml and *.json).")
			return comps, cobra.ShellCompDirectiveFilterFileExt
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return validatecmd.Run(args, cmdFlags)
		},
	}
	cmd.Flags().StringVarP(&cmdFlags.VarsFilePath, "vars", "f", "", "Path to the `FILE` containing template variables in JSON or YAML format.")
	cmd.Flags().StringSliceVar(&cmdFlags.Vars, "var", nil, "Template variable in `KEY=VALUE` form. Can be repeated.")

	return cmd
}
