package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/debug/tidy"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func Debug(parentFlags *root.Flags) *cli.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	flags := &flag.Debug{
		Flags:     parentFlags,
		Context:   defaultContext,
		Namespace: defaultNamespace,
	}

	return cli.NewCommand("debug", "Create and attach to a debug container.",
		cli.WithLongDescription(`Create and attach to a debug container

When flag "--copy" is set, the command can be used to debug a copy of the original pod, allowing you to troubleshoot without affecting the live pod.

To debug a live pod, run the command without the "--copy" flag.

You can only reconnect to the debug session if the pod is running.`),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			return debug.Run(args[0], flags)
		}),
		cli.WithStickyFlag("context", "c", "The kubeconfig `CONTEXT` to use. Defaults to current context.", &flags.Context),
		cli.WithStickyFlag("namespace", "n", "The kubernetes `NAMESPACE` to use. Defaults to current namespace.", &flags.Namespace),
		cli.WithStickyFlag("copy", "", "Create a copy of the pod with a debug container. The original pod remains running and unaffected.", &flags.Copy),
		cli.WithFlag("by-pod", "b", "Attach to a specific `BY-POD` in a workload.", &flags.ByPod),
		cli.WithSubCommands(tidyCommand(flags)),
	)
}

func tidyCommand(parentFlags *flag.Debug) *cli.Command {
	return cli.NewCommand("tidy", "Clean up debug containers and debug pods.",
		cli.WithLongDescription(`Remove debug containers created by the "nais debug" command

Set the "--copy" flag to delete copy pods.`),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			return tidy.Run(args[0], &flag.DebugTidy{Debug: parentFlags})
		}),
	)
}
