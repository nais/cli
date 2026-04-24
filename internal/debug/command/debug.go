package command

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func Debug(parentFlags *flags.GlobalFlags) *naistrix.Command {
	stickyFlags := &flag.DebugSticky{
		GlobalFlags: parentFlags,
		Environment: flag.Environment(""),
	}

	debugFlags := &flag.Debug{
		DebugSticky: stickyFlags,
	}

	return &naistrix.Command{
		Name:  "debug",
		Title: "Create and attach to a debug container.",
		Description: heredoc.Doc(`
			When flag "--copy" is set, the command can be used to debug a copy of the original pod, allowing you to troubleshoot without affecting the live pod.

			To debug a live pod, run the command without the "--copy" flag.

			You can only reconnect to the debug session if the pod is running.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags:       debugFlags,
		StickyFlags: stickyFlags,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := stickyFlags.UsesRemovedFlags(); err != nil {
				return err
			}
			if _, err := debugFlags.RequiredTeam(); err != nil {
				return err
			}
			if debugFlags.Environment == "" {
				return fmt.Errorf("the -e, --environment flag is required")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return debug.Run(args.Get("app_name"), debugFlags, out)
		},
	}
}
