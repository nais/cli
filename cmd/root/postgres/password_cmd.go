package postgres

import (
	"fmt"
	"github.com/spf13/cobra"
)

var passwordCmd = &cobra.Command{
	Use:     "password [command]",
	Short:   "Administrate Postgres password",
	Example: `nais postgres password rotate`,
	RunE: func(command *cobra.Command, args []string) error {
		return fmt.Errorf("missing required command")
	},
}
