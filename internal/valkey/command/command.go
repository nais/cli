package command

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func Valkey(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.Valkey{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "valkey",
		Title:       "Manage Valkey instances.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			createValkey(flags),
			deleteValkey(flags),
			describeValkey(flags),
			listValkeys(flags),
			updateValkey(flags),
		},
	}
}

var (
	// TODO(tronghn): `team` and `environment` are currently required arguments for many of these subcommands.
	//  These should be re-usable configuration options (e.g. a 'shared' package for consistency) with command-specific
	//  flags to override per invocation instead of requiring repeated arguments.
	defaultArgs = []naistrix.Argument{
		{Name: "team"},
		{Name: "environment"},
		{Name: "name"},
	}
	defaultValidateFunc = func(_ context.Context, args *naistrix.Arguments) error {
		if args.Len() != 3 {
			return fmt.Errorf("expected 3 arguments, got %d", args.Len())
		}
		if args.Get("team") == "" {
			return fmt.Errorf("team cannot be empty")
		}
		if args.Get("environment") == "" {
			return fmt.Errorf("environment cannot be empty")
		}
		if args.Get("name") == "" {
			return fmt.Errorf("name cannot be empty")
		}
		return nil
	}
)

func metadataFromArgs(args *naistrix.Arguments) valkey.Metadata {
	return valkey.Metadata{
		TeamSlug:        args.Get("team"),
		EnvironmentName: args.Get("environment"),
		Name:            args.Get("name"),
	}
}
