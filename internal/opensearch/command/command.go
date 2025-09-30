package command

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
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

func metadataFromArgs(args []string) opensearch.Metadata {
	return opensearch.Metadata{
		TeamSlug:        args[0],
		EnvironmentName: args[1],
		Name:            args[2],
	}
}

func normalizeStorage(tier gql.OpenSearchTier, memory gql.OpenSearchMemory, storage int) (int, error) {
	memories, ok := storageRanges[tier]
	if !ok {
		return 0, fmt.Errorf("invalid OpenSearch tier: %s", tier)
	}

	srange, ok := memories[memory]
	if !ok {
		return 0, fmt.Errorf("invalid OpenSearch memory for tier. %v cannot have memory %v", tier, memory)
	}

	// If storage is not specified, use the minimum for the given tier and memory.
	if storage <= 0 {
		return srange.Min, nil
	}

	if storage < srange.Min || storage > srange.Max {
		return 0, fmt.Errorf("invalid storage capacity %d for tier %s with memory %s, must be between %d and %d", storage, tier, memory, srange.Min, srange.Max)
	}

	return storage, nil
}

type storageRange struct {
	Min int
	Max int
}

// TODO: this should probably be returned by the API instead of being hard coded here?
var storageRanges = map[gql.OpenSearchTier]map[gql.OpenSearchMemory]storageRange{
	gql.OpenSearchTierSingleNode: {
		gql.OpenSearchMemoryGb2:  {Min: 16, Max: 16},
		gql.OpenSearchMemoryGb4:  {Min: 80, Max: 400},
		gql.OpenSearchMemoryGb8:  {Min: 175, Max: 875},
		gql.OpenSearchMemoryGb16: {Min: 350, Max: 1750},
		gql.OpenSearchMemoryGb32: {Min: 700, Max: 3500},
		gql.OpenSearchMemoryGb64: {Min: 1400, Max: 5120},
	},
	gql.OpenSearchTierHighAvailability: {
		gql.OpenSearchMemoryGb4:  {Min: 240, Max: 1200},
		gql.OpenSearchMemoryGb8:  {Min: 525, Max: 2625},
		gql.OpenSearchMemoryGb16: {Min: 1050, Max: 5250},
		gql.OpenSearchMemoryGb32: {Min: 2100, Max: 10500},
		gql.OpenSearchMemoryGb64: {Min: 4200, Max: 15360},
	},
}
