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
	return &cli.Command{
		Name:  "validate",
		Short: "Validate one or more Nais manifest files.",
		Args: []cli.Argument{
			{Name: "file", Repeatable: true},
		},
		ValidateFunc:           cli.ValidateMinArgs(1),
		AutoCompleteExtensions: []string{"yaml", "yml", "json"},
		Flags:                  flags,
		RunFunc: func(ctx context.Context, out output.Output, args []string) error {
			return validate.Run(args, flags, out)
		},
	}
}
