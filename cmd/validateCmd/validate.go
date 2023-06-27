package validateCmd

import (
	"fmt"
	"github.com/urfave/cli/v2"

	"github.com/nais/cli/pkg/validate"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "validate",
		Aliases:         []string{"v"},
		Description:     "Validate nais.yaml configuration.",
		HideHelpCommand: true,
		Action: func(context *cli.Context) error {
			if context.Args().Len() == 0 {
				return fmt.Errorf("no config files provided")
			}

			return validate.NaisConfig(context.Args().Slice())
		},
	}
}
