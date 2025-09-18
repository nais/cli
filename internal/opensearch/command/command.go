package command

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
)

func OpenSearch(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.OpenSearch{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "opensearch",
		Title:       "Manage OpenSearch instances.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			createOpenSearch(flags),
			deleteOpenSearch(flags),
			describeOpenSearch(flags),
			listOpenSearches(flags),
			updateOpenSearch(flags),
		},
	}
}

var (
	// TODO(tronghn): team and environment should be a global configuration option with global flags to override
	//  per invocation instead of arguments.
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

func metadataFromArgs(args []string) opensearch.Metadata {
	return opensearch.Metadata{
		TeamSlug:        args[0],
		EnvironmentName: args[1],
		Name:            args[2],
	}
}
