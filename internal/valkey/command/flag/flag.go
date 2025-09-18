package flag

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Valkey struct {
	*alpha.Alpha
}

type Create struct {
	*Valkey
	Size            Size            `name:"size" short:"s" usage:"|SIZE| of the Valkey instance. Defaults to |RAM_1GB|."`
	Tier            Tier            `name:"tier" short:"t" usage:"|TIER| of the Valkey instance. Defaults to |HIGH_AVAILABILITY|."`
	MaxMemoryPolicy MaxMemoryPolicy `name:"max-memory-policy" short:"m" usage:"|MAX_MEMORY_POLICY| for the Valkey instance. Defaults to |NO_EVICTION|."`
}

func (c *Create) Validate() error {
	if c.Size != "" && !c.Size.IsValid() {
		return fmt.Errorf("invalid size %q, must be one of: %v", c.Size, gql.AllValkeySize)
	}
	if c.Tier != "" && !c.Tier.IsValid() {
		return fmt.Errorf("invalid tier %q, must be one of: %v", c.Tier, gql.AllValkeyTier)
	}
	if c.MaxMemoryPolicy != "" && !c.MaxMemoryPolicy.IsValid() {
		return fmt.Errorf("invalid max memory policy %q, must be one of: %v", c.MaxMemoryPolicy, gql.AllValkeyMaxMemoryPolicy)
	}
	return nil
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

type Update struct {
	*Valkey
	Size            Size            `name:"size" short:"s" usage:"|SIZE| of the Valkey instance."`
	Tier            Tier            `name:"tier" short:"t" usage:"|TIER| of the Valkey instance."`
	MaxMemoryPolicy MaxMemoryPolicy `name:"max-memory-policy" short:"m" usage:"|MAX_MEMORY_POLICY| for the Valkey instance."`
}

func (u *Update) Validate() error {
	if u.Size != "" && !u.Size.IsValid() {
		return fmt.Errorf("invalid size %q, must be one of: %v", u.Size, gql.AllValkeySize)
	}
	if u.Tier != "" && !u.Tier.IsValid() {
		return fmt.Errorf("invalid tier %q, must be one of: %v", u.Tier, gql.AllValkeyTier)
	}
	if u.MaxMemoryPolicy != "" && !u.MaxMemoryPolicy.IsValid() {
		return fmt.Errorf("invalid max memory policy %q, must be one of: %v", u.MaxMemoryPolicy, gql.AllValkeyMaxMemoryPolicy)
	}
	return nil
}

type Size string

func (u *Size) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	sizes := make([]string, 0, len(gql.AllValkeySize))
	for _, size := range gql.AllValkeySize {
		sizes = append(sizes, string(size))
	}
	return sizes, "Available Valkey sizes."
}

func (u *Size) IsValid() bool {
	for _, size := range gql.AllValkeySize {
		if string(size) == string(*u) {
			return true
		}
	}
	return false
}

type Tier string

func (t *Tier) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	tiers := make([]string, 0, len(gql.AllValkeyTier))
	for _, tier := range gql.AllValkeyTier {
		tiers = append(tiers, string(tier))
	}
	return tiers, "Available Valkey tiers."
}

func (t *Tier) IsValid() bool {
	for _, tier := range gql.AllValkeyTier {
		if string(tier) == string(*t) {
			return true
		}
	}
	return false
}

type MaxMemoryPolicy string

func (m *MaxMemoryPolicy) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	policies := make([]string, 0, len(gql.AllValkeyMaxMemoryPolicy))
	for _, policy := range gql.AllValkeyMaxMemoryPolicy {
		policies = append(policies, string(policy))
	}
	return policies, "Available Valkey max memory policies."
}

func (m *MaxMemoryPolicy) IsValid() bool {
	for _, policy := range gql.AllValkeyMaxMemoryPolicy {
		if string(policy) == string(*m) {
			return true
		}
	}
	return false
}
