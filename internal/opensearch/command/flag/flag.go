package flag

import (
	"context"
	"fmt"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

type OpenSearch struct {
	*alpha.Alpha
}

type Create struct {
	*OpenSearch
	Memory    Memory  `name:"memory" short:"m" usage:"|MEMORY| of the OpenSearch instance. Defaults to |GB_4|."`
	Tier      Tier    `name:"tier" short:"t" usage:"|TIER| of the OpenSearch instance. Defaults to |SINGLE_NODE|."`
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
		return fmt.Errorf("invalid major version %q, must be one of: %v", c.Version, gql.AllOpenSearchMajorVersion)
	}
	return nil
}

type Delete struct {
	*OpenSearch
}

type Describe struct {
	*OpenSearch
}

type List struct {
	*OpenSearch
}

type Update struct {
	*OpenSearch
	Memory       Memory  `name:"memory" short:"m" usage:"|MEMORY| of the OpenSearch instance."`
	Tier         Tier    `name:"tier" short:"t" usage:"|TIER| of the OpenSearch instance."`
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
		return fmt.Errorf("invalid major version %q, must be one of: %v", u.MajorVersion, gql.AllOpenSearchMajorVersion)
	}
	return nil
}

type Memory string

func (s *Memory) AutoComplete(context.Context, []string, string, any) ([]string, string) {
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

func (t *Tier) AutoComplete(context.Context, []string, string, any) ([]string, string) {
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

func (v *Version) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	versions := make([]string, 0, len(gql.AllOpenSearchMajorVersion))
	for _, version := range gql.AllOpenSearchMajorVersion {
		versions = append(versions, string(version))
	}
	return versions, "Available OpenSearch major versions."
}

func (v *Version) IsValid() bool {
	for _, version := range gql.AllOpenSearchMajorVersion {
		if string(version) == string(*v) {
			return true
		}
	}
	return false
}
