package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/cli/v2/internal/validate"
	"github.com/nais/cli/v2/internal/validate/command/flag"
)

func Validate(parentFlags *root.Flags) *cli.Command {
	flags := &flag.Validate{Flags: parentFlags}
	return &cli.Command{
		Name:  "validate",
		Title: "Validate one or more Nais manifest files.",
		Args: []cli.Argument{
			{Name: "file", Repeatable: true},
		},
		AutoCompleteExtensions: []string{"yaml", "yml", "json"},
		Flags:                  flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return validate.Run(args, flags, out)
		},
	}
}
