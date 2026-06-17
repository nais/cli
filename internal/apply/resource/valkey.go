package resource

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
	"gopkg.in/yaml.v3"
)

func init() {
	register(valkeyResource{kindSupport{kind: "Valkey", strippedVersion: "v1"}})
}

// valkeyResource applies Valkey instances through the nais-api Valkey mutation.
// It is only available as a stripped manifest; a raw Valkey CRD falls through to
// the generic apply endpoint.
type valkeyResource struct{ kindSupport }

// valkeySpec is the user-facing (CRD-flavoured) Valkey spec. Values use the CRD
// vocabulary (e.g. "4GB", "HighAvailability") and are mapped to GraphQL enums
// before the mutation is called.
type valkeySpec struct {
	Memory               string             `yaml:"memory"`
	Tier                 string             `yaml:"tier"`
	MaxMemoryPolicy      string             `yaml:"maxMemoryPolicy"`
	Databases            int                `yaml:"databases"`
	NotifyKeyspaceEvents string             `yaml:"notifyKeyspaceEvents"`
	Persistence          *valkeyPersistence `yaml:"persistence"`
}

// valkeyPersistence is parsed so it does not trip the strict spec decoder, but
// is ignored until the nais-api mutation supports it.
type valkeyPersistence struct {
	Disabled bool `yaml:"disabled"`
}

var (
	valkeyMemory = map[string]gql.ValkeyMemory{
		"1GB":   gql.ValkeyMemoryGb1,
		"4GB":   gql.ValkeyMemoryGb4,
		"8GB":   gql.ValkeyMemoryGb8,
		"14GB":  gql.ValkeyMemoryGb14,
		"28GB":  gql.ValkeyMemoryGb28,
		"56GB":  gql.ValkeyMemoryGb56,
		"112GB": gql.ValkeyMemoryGb112,
		"200GB": gql.ValkeyMemoryGb200,
	}
	valkeyTier = map[string]gql.ValkeyTier{
		"SingleNode":       gql.ValkeyTierSingleNode,
		"HighAvailability": gql.ValkeyTierHighAvailability,
	}
	valkeyMaxMemoryPolicy = map[string]gql.ValkeyMaxMemoryPolicy{
		"allkeys-lfu":     gql.ValkeyMaxMemoryPolicyAllkeysLfu,
		"allkeys-lru":     gql.ValkeyMaxMemoryPolicyAllkeysLru,
		"allkeys-random":  gql.ValkeyMaxMemoryPolicyAllkeysRandom,
		"noeviction":      gql.ValkeyMaxMemoryPolicyNoEviction,
		"volatile-lfu":    gql.ValkeyMaxMemoryPolicyVolatileLfu,
		"volatile-lru":    gql.ValkeyMaxMemoryPolicyVolatileLru,
		"volatile-random": gql.ValkeyMaxMemoryPolicyVolatileRandom,
		"volatile-ttl":    gql.ValkeyMaxMemoryPolicyVolatileTtl,
	}
)

func (v valkeyResource) Apply(ctx context.Context, meta Metadata, spec *yaml.Node) (Action, error) {
	var s valkeySpec
	if err := decodeSpec(spec, &s); err != nil {
		return "", err
	}

	data := &valkey.Valkey{
		Databases:            s.Databases,
		NotifyKeyspaceEvents: s.NotifyKeyspaceEvents,
		Labels:               meta.Labels,
	}

	var err error
	if data.Memory, err = enumValue("memory", s.Memory, valkeyMemory); err != nil {
		return "", err
	}
	if data.Tier, err = enumValue("tier", s.Tier, valkeyTier); err != nil {
		return "", err
	}
	if s.MaxMemoryPolicy != "" {
		if data.MaxMemoryPolicy, err = enumValue("maxMemoryPolicy", s.MaxMemoryPolicy, valkeyMaxMemoryPolicy); err != nil {
			return "", err
		}
	}
	// s.Persistence is parsed but intentionally ignored until the API supports it.

	vmeta := valkey.Metadata{
		Name:            meta.Name,
		EnvironmentName: meta.EnvironmentName,
		TeamSlug:        meta.TeamSlug,
	}

	exists, err := v.exists(ctx, vmeta)
	if err != nil {
		return "", err
	}
	if exists {
		if _, err := valkey.Update(ctx, vmeta, data); err != nil {
			return "", err
		}
		return ActionUpdated, nil
	}
	if _, err := valkey.Create(ctx, vmeta, data); err != nil {
		return "", err
	}
	return ActionCreated, nil
}

// exists reports whether a Valkey instance with the given name already exists.
func (v valkeyResource) exists(ctx context.Context, meta valkey.Metadata) (bool, error) {
	_, err := valkey.Get(ctx, meta)
	if err != nil {
		if naisapi.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
