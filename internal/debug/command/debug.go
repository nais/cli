package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/naistrix"
)

func Debug(parentFlags *naistrix.GlobalFlags) *naistrix.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	stickyFlags := &flag.DebugSticky{
		GlobalFlags: parentFlags,
		Context:     flag.Context(defaultContext),
		Namespace:   defaultNamespace,
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
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return debug.Run(args.Get("app_name"), debugFlags)
		},
	}
}
