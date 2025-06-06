package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/debug/tidy"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/root"
)

func Debug(parentFlags *root.Flags) *cli.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	stickyFlags := &flag.DebugSticky{
		Flags:     parentFlags,
		Context:   defaultContext,
		Namespace: defaultNamespace,
	}

	debugFlags := &flag.Debug{
		DebugSticky: stickyFlags,
	}

	return &cli.Command{
		Name:  "debug",
		Short: "Create and attach to a debug container.",
		Long: `Create and attach to a debug container

When flag "--copy" is set, the command can be used to debug a copy of the original pod, allowing you to troubleshoot without affecting the live pod.

To debug a live pod, run the command without the "--copy" flag.

You can only reconnect to the debug session if the pod is running.`,
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
		},
		Flags:        debugFlags,
		StickyFlags:  stickyFlags,
		ValidateFunc: cli.ValidateMinArgs(1),
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return debug.Run(args[0], debugFlags)
		},
		SubCommands: []*cli.Command{
			tidyCommand(stickyFlags),
		},
	}
}

func tidyCommand(parentFlags *flag.DebugSticky) *cli.Command {
	flags := &flag.DebugTidy{
		DebugSticky: parentFlags,
	}
	return &cli.Command{
		Name:  "tidy",
		Short: "Clean up debug containers and debug pods.",

		Long: `Remove debug containers created by the "nais debug" command

Set the "--copy" flag to delete copy pods.`,
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		Flags:        flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return tidy.Run(args[0], flags)
		},
	}
}
