package command

import (
	"context"
	"fmt"
	"os"
	"sort"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func OpenSearch(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.OpenSearch{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "opensearch",
		Aliases:     []string{"opensearches", "os"},
		Title:       "Manage OpenSearch instances.",
		StickyFlags: flags,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		SubCommands: []*naistrix.Command{
			create(flags),
			delete(flags),
			get(flags),
			list(flags),
			update(flags),
		},
	}
}

func validateArgs(args *naistrix.Arguments) error {
	if args.Len() != 1 {
		return fmt.Errorf("expected 1 arguments, got %d", args.Len())
	}
	if args.Get("name") == "" {
		return fmt.Errorf("name cannot be empty")
	}

	return nil
}

func metadataFromArgs(args *naistrix.Arguments, team string, environment string) opensearch.Metadata {
	return opensearch.Metadata{
		TeamSlug:        team,
		EnvironmentName: environment,
		Name:            args.Get("name"),
	}
}

func autoCompleteOpenSearchNames(ctx context.Context, team, environment string, requireEnvironment bool) ([]string, string) {
	if team == "" {
		return nil, "Please provide team to auto-complete OpenSearch instance names. 'nais config set team <team>', or '--team <team>' flag."
	}

	if environment == "" {
		envs := environmentValuesFromCLIArgs()
		if len(envs) > 1 {
			return nil, "Please specify exactly one environment to auto-complete OpenSearch instance names. '--environment <environment>' flag."
		}
		if len(envs) == 1 {
			environment = envs[0]
		}
	}

	if requireEnvironment && environment == "" {
		return nil, "Please provide environment to auto-complete OpenSearch instance names. '--environment <environment>' flag."
	}

	instances, err := opensearch.GetAll(ctx, team)
	if err != nil {
		return nil, "Unable to fetch OpenSearch instances."
	}

	seen := make(map[string]struct{})
	var names []string
	for _, instance := range instances {
		if environment != "" && instance.TeamEnvironment.Environment.Name != environment {
			continue
		}
		if _, ok := seen[instance.Name]; ok {
			continue
		}
		seen[instance.Name] = struct{}{}
		names = append(names, instance.Name)
	}

	sort.Strings(names)
	if len(names) == 0 && environment != "" {
		return nil, fmt.Sprintf("No OpenSearch instances found in environment %q.", environment)
	}

	return names, "Select an OpenSearch instance."
}

func environmentValuesFromCLIArgs() []string {
	return cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
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
