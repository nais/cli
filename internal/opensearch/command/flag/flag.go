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
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/naistrix"
)

type OpenSearch struct {
	*flags.GlobalFlags
	Environment Env `name:"environment" short:"e" usage:"Filter by environment."`
}

type Env string

func (e *Env) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	var team string
	switch f := flags.(type) {
	case *Credentials:
		team = f.Team
	case *OpenSearch:
		team = f.Team
	}

	if team != "" && isCredentialsCompletionFromCLIArgs() {
		envs, err := opensearchCredentialEnvironments(ctx, team)
		if err == nil {
			return envs, "Available environments with OpenSearch instances"
		}
	}
	return autoCompleteEnvironments(ctx)
}

func isCredentialsCompletionFromCLIArgs() bool {
	return slices.Contains(os.Args, "credentials")
}

func opensearchCredentialEnvironments(ctx context.Context, team string) ([]string, error) {
	instances, err := opensearch.GetAll(ctx, team)
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

type GetEnv string

func (e *GetEnv) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	if args.Len() == 0 {
		return autoCompleteEnvironments(ctx)
	}

	f := flags.(*Get)
	if len(f.Team) == 0 {
		return nil, "Please provide team to auto-complete environments. 'nais config team set <team>', or '--team <team>' flag."
	}

	envs, err := opensearch.OpenSearchEnvironments(ctx, f.Team, args.Get("name"))
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Get struct {
	*OpenSearch
	Environment GetEnv `name:"environment" short:"e" usage:"Filter by environment."`
}

type Output string

type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	return autoCompleteEnvironments(ctx)
}

type List struct {
	*OpenSearch
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
	Output      Output       `name:"output" short:"o" usage:"Format output (table|json)."`
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
	perms := make([]string, 0, len(gql.AllAivenPermission))
	for _, perm := range gql.AllAivenPermission {
		perms = append(perms, string(perm))
	}
	return perms, "Available permission levels."
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}
