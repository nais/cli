package flag

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/naistrix"
)

type Environments []string

type Valkey struct {
	*flags.GlobalFlags
	Environment Env `name:"environment" short:"e" usage:"Filter by environment."`
}

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	return autoCompleteEnvironments(ctx)
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Create struct {
	*Valkey
	Yes             bool            `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
	Memory          Memory          `name:"memory" short:"m" usage:"|MEMORY| of the Valkey instance. Defaults to |GB_1|."`
	Tier            Tier            `name:"tier" usage:"|TIER| of the Valkey instance. Defaults to |HIGH_AVAILABILITY|."`
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

type Output string

type (
	Env  string
	List struct {
		*Valkey
		Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
		Output      Output       `name:"output" short:"o" usage:"Format output (table|json)."`
	}
)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

func (e *Env) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	var team string
	switch f := flags.(type) {
	case *Credentials:
		team = f.Team
	case *Valkey:
		team = f.Team
	}

	if team != "" && isCredentialsCompletionFromCLIArgs() {
		envs, err := valkeyCredentialEnvironments(ctx, team)
		if err == nil {
			return envs, "Available environments with Valkey instances"
		}
	}
	return autoCompleteEnvironments(ctx)
}

func isCredentialsCompletionFromCLIArgs() bool {
	return slices.Contains(os.Args, "credentials")
}

func valkeyCredentialEnvironments(ctx context.Context, team string) ([]string, error) {
	instances, err := valkey.GetAll(ctx, team)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var envs []string
	for _, instance := range instances {
		env := instance.TeamEnvironment.Environment.Name
		if _, ok := seen[env]; ok {
			continue
		}
		seen[env] = struct{}{}
		envs = append(envs, env)
	}

	sort.Strings(envs)
	return envs, nil
}

type Update struct {
	*Valkey
	Yes             bool            `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
	Memory          Memory          `name:"memory" short:"m" usage:"|MEMORY| of the Valkey instance."`
	Tier            Tier            `name:"tier" usage:"|TIER| of the Valkey instance."`
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

type Credentials struct {
	*Valkey
	Permission Permission `name:"permission" short:"p" usage:"Permission level for the credentials (READ, WRITE, READWRITE, ADMIN)."`
	TTL        string     `name:"ttl" usage:"Time-to-live for the credentials (e.g. '1d', '7d'). Maximum 30 days."`
}

type Permission string

var _ naistrix.FlagAutoCompleter = (*Permission)(nil)

func (p *Permission) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	perms := make([]string, 0, len(gql.AllCredentialPermission))
	for _, perm := range gql.AllCredentialPermission {
		perms = append(perms, string(perm))
	}
	return perms, "Available permission levels."
}
