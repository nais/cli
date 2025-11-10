package command

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validate"
	"github.com/nais/cli/internal/validate/command/flag"
	"github.com/nais/naistrix"
)

func Validate(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Validate{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:  "validate",
		Title: "Validate one or more Nais manifest files.",
		Args: []naistrix.Argument{
			{Name: "file", Repeatable: true},
		},
		AutoCompleteExtensions: []string{"yaml", "yml", "json"},
		Flags:                  flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return validate.Run(args.All(), flags, out)
		},
	}
}
