package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type OpenSearch struct {
	*flags.GlobalFlags
}

type Create struct {
	*OpenSearch
	Memory    Memory  `name:"memory" short:"m" usage:"|MEMORY| of the OpenSearch instance. Defaults to |GB_4|."`
	Tier      Tier    `name:"tier" usage:"|TIER| of the OpenSearch instance. Defaults to |SINGLE_NODE|."`
	Version   Version `name:"version" usage:"Major |VERSION| of the OpenSearch instance. Defaults to |V2|."`
	StorageGB int     `name:"storage-gb" usage:"Storage capacity in |GB| for the OpenSearch instance. Defaults vary for different combinations of |TIER| and |MEMORY|."`
}

func (c *Create) Validate() error {
	if c.Memory != "" && !c.Memory.IsValid() {
		return fmt.Errorf("invalid memory %q, must be one of: %v", c.Memory, gql.AllOpenSearchMemory)
	}
	if c.Tier != "" && !c.Tier.IsValid() {
		return fmt.Errorf("invalid tier %q, must be one of: %v", c.Tier, gql.AllOpenSearchTier)
	}
	if c.Version != "" && !c.Version.IsValid() {
		return fmt.Errorf("invalid version %q, must be one of: %v", c.Version, gql.AllOpenSearchMajorVersion)
	}
	return nil
}

type Delete struct {
	*OpenSearch
}

type Get struct {
	*OpenSearch
}

type Output string

type List struct {
	*OpenSearch
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
}

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Update struct {
	*OpenSearch
	Memory       Memory  `name:"memory" short:"m" usage:"|MEMORY| of the OpenSearch instance."`
	Tier         Tier    `name:"tier" usage:"|TIER| of the OpenSearch instance."`
	MajorVersion Version `name:"version" usage:"Major |VERSION| of the OpenSearch instance."`
	StorageGB    int     `name:"storage-gb" usage:"Storage capacity in |GB| for the OpenSearch instance. Defaults vary for different combinations of |TIER| and |MEMORY|."`
}

func (u *Update) Validate() error {
	if u.Memory != "" && !u.Memory.IsValid() {
		return fmt.Errorf("invalid memory %q, must be one of: %v", u.Memory, gql.AllOpenSearchMemory)
	}
	if u.Tier != "" && !u.Tier.IsValid() {
		return fmt.Errorf("invalid tier %q, must be one of: %v", u.Tier, gql.AllOpenSearchTier)
	}
	if u.MajorVersion != "" && !u.MajorVersion.IsValid() {
		return fmt.Errorf("invalid version %q, must be one of: %v", u.MajorVersion, gql.AllOpenSearchMajorVersion)
	}
	return nil
}

type Memory string

var _ naistrix.FlagAutoCompleter = (*Memory)(nil)

func (s *Memory) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	memories := make([]string, 0, len(gql.AllOpenSearchMemory))
	for _, memory := range gql.AllOpenSearchMemory {
		memories = append(memories, string(memory))
	}
	return memories, "Available OpenSearch memory values."
}

func (s *Memory) IsValid() bool {
	for _, memory := range gql.AllOpenSearchMemory {
		if string(memory) == string(*s) {
			return true
		}
	}
	return false
}

type Tier string

var _ naistrix.FlagAutoCompleter = (*Tier)(nil)

func (t *Tier) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	tiers := make([]string, 0, len(gql.AllOpenSearchTier))
	for _, tier := range gql.AllOpenSearchTier {
		tiers = append(tiers, string(tier))
	}
	return tiers, "Available OpenSearch tiers."
}

func (t *Tier) IsValid() bool {
	for _, tier := range gql.AllOpenSearchTier {
		if string(tier) == string(*t) {
			return true
		}
	}
	return false
}

type Version string

var _ naistrix.FlagAutoCompleter = (*Version)(nil)

func (v *Version) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	versions := make([]string, 0, len(gql.AllOpenSearchMajorVersion))
	for _, version := range gql.AllOpenSearchMajorVersion {
		versions = append(versions, string(version))
	}
	return versions, "Available OpenSearch versions."
}

func (v *Version) IsValid() bool {
	for _, version := range gql.AllOpenSearchMajorVersion {
		if string(version) == string(*v) {
			return true
		}
	}
	return false
}

type Credentials struct {
	*OpenSearch
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
