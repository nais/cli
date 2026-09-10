package flag

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"k8s.io/apimachinery/pkg/api/resource"
)

type OpenSearch struct {
	*flags.GlobalFlags
}

type Create struct {
	*OpenSearch
	Memory                         Memory  `name:"memory" short:"m" usage:"|MEMORY| of the OpenSearch instance. Defaults to |GB_4|."`
	Tier                           Tier    `name:"tier" usage:"|TIER| of the OpenSearch instance. Defaults to |SINGLE_NODE|."`
	Version                        Version `name:"version" usage:"Major |VERSION| of the OpenSearch instance. Defaults to |V2|."`
	StorageGB                      int     `name:"storage-gb" usage:"Storage capacity in |GB| for the OpenSearch instance. Defaults vary for different combinations of |TIER| and |MEMORY|."`
	ShardIndexingPressureEnabled   Boolean `name:"shard-indexing-pressure-enabled" usage:"Enable shard indexing pressure (true or false)."`
	ShardIndexingPressureEnforced  Boolean `name:"shard-indexing-pressure-enforced" usage:"Enforce shard indexing pressure limits (true or false)."`
	IndicesQueryBoolMaxClauseCount int     `name:"indices-query-bool-max-clause-count" usage:"Maximum number of clauses in a Lucene BooleanQuery."`
	HttpMaxContentLength           string  `name:"http-max-content-length" usage:"Maximum HTTP request content length (for example, 100Mi)."`
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
	if _, err := c.ShardIndexingPressureEnabled.Bool(); err != nil {
		return fmt.Errorf("invalid shard indexing pressure enabled: %w", err)
	}
	if _, err := c.ShardIndexingPressureEnforced.Bool(); err != nil {
		return fmt.Errorf("invalid shard indexing pressure enforced: %w", err)
	}
	if err := validateMaxContentLength(c.HttpMaxContentLength); err != nil {
		return err
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
	Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
	Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
}

func (*List) LabelFacetResource() string { return "openSearches" }

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Update struct {
	*OpenSearch
	Memory                         Memory  `name:"memory" short:"m" usage:"|MEMORY| of the OpenSearch instance."`
	Tier                           Tier    `name:"tier" usage:"|TIER| of the OpenSearch instance."`
	MajorVersion                   Version `name:"version" usage:"Major |VERSION| of the OpenSearch instance."`
	StorageGB                      int     `name:"storage-gb" usage:"Storage capacity in |GB| for the OpenSearch instance. Defaults vary for different combinations of |TIER| and |MEMORY|."`
	ShardIndexingPressureEnabled   Boolean `name:"shard-indexing-pressure-enabled" usage:"Enable shard indexing pressure (true or false)."`
	ShardIndexingPressureEnforced  Boolean `name:"shard-indexing-pressure-enforced" usage:"Enforce shard indexing pressure limits (true or false)."`
	IndicesQueryBoolMaxClauseCount int     `name:"indices-query-bool-max-clause-count" usage:"Maximum number of clauses in a Lucene BooleanQuery."`
	HttpMaxContentLength           string  `name:"http-max-content-length" usage:"Maximum HTTP request content length (for example, 100Mi)."`
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
	if _, err := u.ShardIndexingPressureEnabled.Bool(); err != nil {
		return fmt.Errorf("invalid shard indexing pressure enabled: %w", err)
	}
	if _, err := u.ShardIndexingPressureEnforced.Bool(); err != nil {
		return fmt.Errorf("invalid shard indexing pressure enforced: %w", err)
	}
	if err := validateMaxContentLength(u.HttpMaxContentLength); err != nil {
		return err
	}
	return nil
}

// Boolean supports optional true or false command-line flags.
type Boolean string

var _ naistrix.FlagAutoCompleter = (*Boolean)(nil)

func (b *Boolean) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"true", "false"}, "Boolean value."
}

func (b Boolean) Bool() (*bool, error) {
	if b == "" {
		return nil, nil
	}
	value, err := strconv.ParseBool(string(b))
	if err != nil {
		return nil, fmt.Errorf("%q must be true or false", b)
	}
	return &value, nil
}

func validateMaxContentLength(value string) error {
	if value == "" {
		return nil
	}
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return fmt.Errorf("invalid HTTP max content length %q: %w", value, err)
	}
	if quantity.Sign() <= 0 {
		return fmt.Errorf("invalid HTTP max content length %q: must be positive", value)
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
	perms := make([]string, len(gql.AllOpenSearchPermission))
	for i, perm := range gql.AllOpenSearchPermission {
		perms[i] = string(perm)
	}
	return perms, "Available permission levels."
}
