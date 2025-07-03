package command

import (
	"context"

	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/cli/v2/internal/validate"
	"github.com/nais/cli/v2/internal/validate/command/flag"
	"github.com/nais/naistrix"
)

func Validate(parentFlags *root.Flags) *naistrix.Command {
	flags := &flag.Validate{Flags: parentFlags}
	return &naistrix.Command{
		Name:  "validate",
		Title: "Validate one or more Nais manifest files.",
		Args: []naistrix.Argument{
			{Name: "file", Repeatable: true},
		},
		AutoCompleteExtensions: []string{"yaml", "yml", "json"},
		Flags:                  flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return validate.Run(args, flags, out)
		},
	}
}
