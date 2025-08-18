package command

import (
	"context"

	"github.com/nais/cli/internal/apply"
	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Apply(parentFlags *root.Flags) *naistrix.Command {
	flags := &flag.Apply{Flags: parentFlags}
	return &naistrix.Command{
		Name:  "apply",
		Title: "Apply resources.",
		Args: []naistrix.Argument{
			{Name: "file", Repeatable: true},
		},
		AutoCompleteExtensions: []string{ /*"yaml", "yml", "json",*/ "toml"},
		Flags:                  flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return apply.Run(ctx, args, flags, out)
		},
	}
}
