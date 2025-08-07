package command

import (
	"context"

	initcmd "github.com/nais/cli/internal/init"
	"github.com/nais/cli/internal/init/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Init(parentFlags *root.Flags) *naistrix.Command {
	flags := &flag.Init{Flags: parentFlags}
	return &naistrix.Command{
		Name:  "init",
		Title: "Generates template files for CI and Workload.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return initcmd.Run(flags, out)
		},
	}
}
