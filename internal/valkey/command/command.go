package command

import (
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

var defaultArgs = []naistrix.Argument{
	{Name: "environment"},
	{Name: "name"},
}

func validateArgs(args *naistrix.Arguments) error {
	if args.Len() != 2 {
		return fmt.Errorf("expected 2 arguments, got %d", args.Len())
	}
	if args.Get("environment") == "" {
		return fmt.Errorf("environment cannot be empty")
	}
	if args.Get("name") == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func metadataFromArgs(args *naistrix.Arguments, team string) valkey.Metadata {
	return valkey.Metadata{
		TeamSlug:        team,
		EnvironmentName: args.Get("environment"),
		Name:            args.Get("name"),
	}
}
