package flag

import (
	"context"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Valkey struct {
	*alpha.Alpha
}

type Delete struct {
	*Valkey
}

type Describe struct {
	*Valkey
}

type List struct {
	*Valkey
}

type Create struct {
	*Valkey
	Size            Size            `name:"size" short:"s" usage:"|SIZE| of the Valkey instance. Defaults to |RAM_1GB|."`
	Tier            Tier            `name:"tier" short:"t" usage:"|TIER| of the Valkey instance. Defaults to |HIGH_AVAILABILITY|."`
	MaxMemoryPolicy MaxMemoryPolicy `name:"max-memory-policy" short:"m" usage:"|MAX_MEMORY_POLICY| for the Valkey instance."`
}

type Update struct {
	*Valkey
	Size            Size            `name:"size" short:"s" usage:"|SIZE| of the Valkey instance."`
	Tier            Tier            `name:"tier" short:"t" usage:"|TIER| of the Valkey instance."`
	MaxMemoryPolicy MaxMemoryPolicy `name:"max-memory-policy" short:"m" usage:"|MAX_MEMORY_POLICY| for the Valkey instance."`
}

type Size string

func (t *Size) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	sizes := make([]string, 0, len(gql.AllValkeySize))
	for _, size := range gql.AllValkeySize {
		sizes = append(sizes, string(size))
	}
	return sizes, "Available Valkey sizes."
}

type Tier string

func (t *Tier) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	tiers := make([]string, 0, len(gql.AllValkeyTier))
	for _, tier := range gql.AllValkeyTier {
		tiers = append(tiers, string(tier))
	}
	return tiers, "Available Valkey tiers."
}

type MaxMemoryPolicy string

func (m *MaxMemoryPolicy) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	policies := make([]string, 0, len(gql.AllValkeyMaxMemoryPolicy))
	for _, policy := range gql.AllValkeyMaxMemoryPolicy {
		policies = append(policies, string(policy))
	}
	return policies, "Available Valkey max memory policies."
}
