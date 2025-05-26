package cli

import (
	"context"

	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

type Application struct {
	cobraCmd *cobra.Command
}

func NewApplication(cmd ...*Command) *Application {
	cc := &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}

	w := output.NewWriter(cc.OutOrStdout())

	for _, c := range cmd {
		c.setup(w)
		cc.AddCommand(c.cobraCmd)
	}

	return &Application{
		cobraCmd: cc,
	}
}

func (a *Application) Run(ctx context.Context) error {
	return a.cobraCmd.ExecuteContext(ctx)
}
