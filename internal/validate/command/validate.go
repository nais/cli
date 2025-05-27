package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/internal/validate"
	"github.com/nais/cli/internal/validate/command/flag"
)

func Validate(parentFlags *root.Flags) *cli.Command {
	flags := &flag.Validate{Flags: parentFlags}
	return cli.NewCommand("validate", "Validate one or more Nais manifest files.",
		cli.WithArgs("file..."),
		cli.WithValidate(cli.ValidateMinArgs(1)),
		cli.WithAutoCompleteFiles("yaml", "yml", "json"),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			return validate.Run(args, flags)
		}),
		cli.WithFlag("vars", "f", "Path to the `FILE` containing template variables in JSON or YAML format.", &flags.VarsFilePath),
		cli.WithFlag("var", "", "Template variable in `KEY=VALUE` form. Can be repeated.", &flags.Vars),
	)
}
