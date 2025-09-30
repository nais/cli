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
	defaultValidateFunc = func(_ context.Context, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("expected 3 arguments, got %d", len(args))
		}
		if args[0] == "" {
			return fmt.Errorf("team cannot be empty")
		}
		if args[1] == "" {
			return fmt.Errorf("environment cannot be empty")
		}
		if args[2] == "" {
			return fmt.Errorf("name cannot be empty")
		}
		return nil
	}
)

func metadataFromArgs(args []string) valkey.Metadata {
	return valkey.Metadata{
		TeamSlug:        args[0],
		EnvironmentName: args[1],
		Name:            args[2],
	}
}
