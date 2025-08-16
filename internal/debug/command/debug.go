package command

import (
	"context"
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/command/flag"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Debug(parentFlags *root.Flags) *naistrix.Command {
	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()
	flags := &flag.Debug{
		Flags:     parentFlags,
		Context:   flag.Context(defaultContext),
		Namespace: defaultNamespace,
		Ttl:       time.Minute,
	}

	return &naistrix.Command{
		Name:  "debug",
		Title: "Create and attach to a debug container.",
		Description: heredoc.Doc(`
			When "--copy" is used the command can be used to debug a copy of the original pod, allowing you to troubleshoot without affecting the live pod.

			To debug a live pod, run the command without the "--copy" flag.
		`),
		ValidateFunc: func(ctx context.Context, args []string) error {
			if flags.Ttl > time.Hour {
				return fmt.Errorf("the --ttl duration can not exceed 1 hour")
			}

			return nil
		},
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return debug.Run(ctx, args[0], flags)
		},
	}
}
