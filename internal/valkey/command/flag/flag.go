package flag

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type Valkey struct {
	*alpha.Alpha
}

type Create struct {
	*Valkey
	Yes             bool            `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
	Memory          Memory          `name:"memory" short:"m" usage:"|MEMORY| of the Valkey instance. Defaults to |GB_1|."`
	Tier            Tier            `name:"tier" short:"t" usage:"|TIER| of the Valkey instance. Defaults to |HIGH_AVAILABILITY|."`
	MaxMemoryPolicy MaxMemoryPolicy `name:"max-memory-policy" usage:"|MAX_MEMORY_POLICY| for the Valkey instance. Defaults to |NO_EVICTION|."`
}

func (c *Create) Validate() error {
	if c.Memory != "" && !c.Memory.IsValid() {
		return fmt.Errorf("invalid memory %q, must be one of: %v", c.Memory, gql.AllValkeyMemory)
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
	Yes bool `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}

type Describe struct {
	*Valkey
}

type List struct {
	*Valkey
}

type Update struct {
	*Valkey
	Yes             bool            `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
	Memory          Memory          `name:"memory" short:"m" usage:"|MEMORY| of the Valkey instance."`
	Tier            Tier            `name:"tier" short:"t" usage:"|TIER| of the Valkey instance."`
	MaxMemoryPolicy MaxMemoryPolicy `name:"max-memory-policy" usage:"|MAX_MEMORY_POLICY| for the Valkey instance."`
}

func (u *Update) Validate() error {
	if u.Memory != "" && !u.Memory.IsValid() {
		return fmt.Errorf("invalid memory %q, must be one of: %v", u.Memory, gql.AllValkeyMemory)
	}
	if u.Tier != "" && !u.Tier.IsValid() {
		return fmt.Errorf("invalid tier %q, must be one of: %v", u.Tier, gql.AllValkeyTier)
	}
	if u.MaxMemoryPolicy != "" && !u.MaxMemoryPolicy.IsValid() {
		return fmt.Errorf("invalid max memory policy %q, must be one of: %v", u.MaxMemoryPolicy, gql.AllValkeyMaxMemoryPolicy)
	}
	return nil
}

type Memory string

var _ naistrix.FlagAutoCompleter = (*Memory)(nil)

func (u *Memory) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	memories := make([]string, 0, len(gql.AllValkeyMemory))
	for _, memory := range gql.AllValkeyMemory {
		memories = append(memories, string(memory))
	}
	return memories, "Available Valkey memory values."
}

func (u *Memory) IsValid() bool {
	for _, memory := range gql.AllValkeyMemory {
		if string(memory) == string(*u) {
			return true
		}
	}
	return false
}

type Tier string

var _ naistrix.FlagAutoCompleter = (*Tier)(nil)

func (t *Tier) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
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

var _ naistrix.FlagAutoCompleter = (*MaxMemoryPolicy)(nil)

func (m *MaxMemoryPolicy) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
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
