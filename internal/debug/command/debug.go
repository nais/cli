package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/v2/internal/debug"
	"github.com/nais/cli/v2/internal/debug/command/flag"
	"github.com/nais/cli/v2/internal/debug/tidy"
	"github.com/nais/cli/v2/internal/k8s"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func Debug(parentFlags *root.Flags) *naistrix.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	stickyFlags := &flag.DebugSticky{
		Flags:     parentFlags,
		Context:   flag.Context(defaultContext),
		Namespace: defaultNamespace,
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
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return debug.Run(args[0], debugFlags)
		},
		SubCommands: []*naistrix.Command{
			tidyCommand(stickyFlags),
		},
	}
}

func tidyCommand(parentFlags *flag.DebugSticky) *naistrix.Command {
	flags := &flag.DebugTidy{
		DebugSticky: parentFlags,
	}
	return &naistrix.Command{
		Name:  "tidy",
		Title: "Clean up debug containers and debug pods.",
		Description: heredoc.Doc(`
			Remove debug containers created by the "nais debug" command.

			Set the "--copy" flag to delete copy pods.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return tidy.Run(args[0], flags)
		},
	}
}
