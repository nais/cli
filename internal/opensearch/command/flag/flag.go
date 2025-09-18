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
	Size    Size    `name:"size" short:"s" usage:"|SIZE| of the OpenSearch instance. Defaults to |RAM_4GB|."`
	Tier    Tier    `name:"tier" short:"t" usage:"|TIER| of the OpenSearch instance. Defaults to |SINGLE_NODE|."`
	Version Version `name:"version" usage:"Major |VERSION| of the OpenSearch instance. Defaults to |V2|."`
}

func (c *Create) Validate() error {
	if c.Size != "" && !c.Size.IsValid() {
		return fmt.Errorf("invalid size %q, must be one of: %v", c.Size, gql.AllOpenSearchSize)
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
	Size         Size    `name:"size" short:"s" usage:"|SIZE| of the OpenSearch instance."`
	Tier         Tier    `name:"tier" short:"t" usage:"|TIER| of the OpenSearch instance."`
	MajorVersion Version `name:"version" short:"m" usage:"Major |VERSION| of the OpenSearch instance."`
}

func (u *Update) Validate() error {
	if u.Size != "" && !u.Size.IsValid() {
		return fmt.Errorf("invalid size %q, must be one of: %v", u.Size, gql.AllOpenSearchSize)
	}
	if u.Tier != "" && !u.Tier.IsValid() {
		return fmt.Errorf("invalid tier %q, must be one of: %v", u.Tier, gql.AllOpenSearchTier)
	}
	if u.MajorVersion != "" && !u.MajorVersion.IsValid() {
		return fmt.Errorf("invalid major version %q, must be one of: %v", u.MajorVersion, gql.AllOpenSearchMajorVersion)
	}
	return nil
}

type Size string

func (s *Size) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	sizes := make([]string, 0, len(gql.AllOpenSearchSize))
	for _, size := range gql.AllOpenSearchSize {
		sizes = append(sizes, string(size))
	}
	return sizes, "Available OpenSearch sizes."
}

func (s *Size) IsValid() bool {
	for _, size := range gql.AllOpenSearchSize {
		if string(size) == string(*s) {
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
