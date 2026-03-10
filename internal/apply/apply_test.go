package apply

import (
	"testing"

	pgratorv1 "github.com/nais/pgrator/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

// ---------------------------------------------------------------------------
// loadManifests
// ---------------------------------------------------------------------------

func TestLoadManifests_Valkey(t *testing.T) {
	objects, err := loadManifests("testdata/nais.yaml")
	require.NoError(t, err)
	require.Len(t, objects, 1)

	v, ok := objects[0].(*pgratorv1.Valkey)
	require.True(t, ok, "expected *pgratorv1.Valkey, got %T", objects[0])
	assert.Equal(t, "foo", v.Name)
	assert.Equal(t, pgratorv1.ValkeyTierSingleNode, v.Spec.Tier)
	assert.Equal(t, pgratorv1.ValkeyMemory4GB, v.Spec.Memory)
	assert.Equal(t, pgratorv1.ValkeyMaxMemoryPolicyAllkeysLRU, v.Spec.MaxMemoryPolicy)
}

func TestLoadManifests_OpenSearch(t *testing.T) {
	objects, err := loadManifests("testdata/opensearch.yaml")
	require.NoError(t, err)
	require.Len(t, objects, 1)

	o, ok := objects[0].(*pgratorv1.OpenSearch)
	require.True(t, ok, "expected *pgratorv1.OpenSearch, got %T", objects[0])
	assert.Equal(t, "myindex", o.Name)
	assert.Equal(t, pgratorv1.OpenSearchTierSingleNode, o.Spec.Tier)
	assert.Equal(t, pgratorv1.OpenSearchMemory4GB, o.Spec.Memory)
	assert.Equal(t, pgratorv1.OpenSearchVersionV2, o.Spec.Version)
	assert.Equal(t, 50, o.Spec.StorageGB)
}

func TestLoadManifests_MultiDocument(t *testing.T) {
	objects, err := loadManifests("testdata/multi.yaml")
	require.NoError(t, err)
	require.Len(t, objects, 2)

	v, ok := objects[0].(*pgratorv1.Valkey)
	require.True(t, ok, "first doc: expected *pgratorv1.Valkey, got %T", objects[0])
	assert.Equal(t, "cache", v.Name)
	assert.Equal(t, pgratorv1.ValkeyTierSingleNode, v.Spec.Tier)
	assert.Equal(t, pgratorv1.ValkeyMemory1GB, v.Spec.Memory)

	o, ok := objects[1].(*pgratorv1.OpenSearch)
	require.True(t, ok, "second doc: expected *pgratorv1.OpenSearch, got %T", objects[1])
	assert.Equal(t, "search", o.Name)
	assert.Equal(t, pgratorv1.OpenSearchTierHighAvailability, o.Spec.Tier)
	assert.Equal(t, pgratorv1.OpenSearchMemory8GB, o.Spec.Memory)
	assert.Equal(t, pgratorv1.OpenSearchVersionV2_19, o.Spec.Version)
	assert.Equal(t, 100, o.Spec.StorageGB)
}

func TestLoadManifests_EmptyPath(t *testing.T) {
	_, err := loadManifests("")
	assert.ErrorContains(t, err, "file path cannot be empty")
}

func TestLoadManifests_UnsupportedExtension(t *testing.T) {
	_, err := loadManifests("testdata/nais.toml")
	assert.ErrorContains(t, err, "unsupported file extension")
}

func TestLoadManifests_MissingFile(t *testing.T) {
	_, err := loadManifests("testdata/nonexistent.yaml")
	assert.ErrorContains(t, err, "failed to read file")
}

func TestLoadManifests_NamespacePreserved(t *testing.T) {
	objects, err := loadManifests("testdata/with-namespace.yaml")
	require.NoError(t, err)
	require.Len(t, objects, 1)

	v, ok := objects[0].(*pgratorv1.Valkey)
	require.True(t, ok, "expected *pgratorv1.Valkey, got %T", objects[0])
	assert.Equal(t, "with-ns", v.Name)
	assert.Equal(t, "my-namespace", v.Namespace, "namespace should be preserved so the warning can fire")
}

// ---------------------------------------------------------------------------
// valkeyFromCRD
// ---------------------------------------------------------------------------

func TestValkeyFromCRD(t *testing.T) {
	for name, tc := range map[string]struct {
		obj     *pgratorv1.Valkey
		wantErr bool
	}{
		"single node no policy": {
			obj: valkeyObj("foo", pgratorv1.ValkeyTierSingleNode, pgratorv1.ValkeyMemory4GB, ""),
		},
		"high availability with policy": {
			obj: valkeyObj("bar", pgratorv1.ValkeyTierHighAvailability, pgratorv1.ValkeyMemory28GB, pgratorv1.ValkeyMaxMemoryPolicyVolatileTTL),
		},
		"bad tier": {
			obj:     valkeyObj("x", "MegaNode", pgratorv1.ValkeyMemory4GB, ""),
			wantErr: true,
		},
		"bad memory": {
			obj:     valkeyObj("x", pgratorv1.ValkeyTierSingleNode, "999GB", ""),
			wantErr: true,
		},
		"bad policy": {
			obj:     valkeyObj("x", pgratorv1.ValkeyTierSingleNode, pgratorv1.ValkeyMemory4GB, "never-evict"),
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := valkeyFromCRD(tc.obj)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestValkeyFromCRD_AllMemorySizes(t *testing.T) {
	memories := []pgratorv1.ValkeyMemory{
		pgratorv1.ValkeyMemory1GB,
		pgratorv1.ValkeyMemory4GB,
		pgratorv1.ValkeyMemory8GB,
		pgratorv1.ValkeyMemory14GB,
		pgratorv1.ValkeyMemory28GB,
		pgratorv1.ValkeyMemory56GB,
		pgratorv1.ValkeyMemory112GB,
		pgratorv1.ValkeyMemory200GB,
	}
	for _, mem := range memories {
		t.Run(string(mem), func(t *testing.T) {
			got, err := valkeyFromCRD(valkeyObj("v", pgratorv1.ValkeyTierSingleNode, mem, ""))
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func TestValkeyFromCRD_AllMaxMemoryPolicies(t *testing.T) {
	policies := []pgratorv1.ValkeyMaxMemoryPolicy{
		pgratorv1.ValkeyMaxMemoryPolicyAllkeysLFU,
		pgratorv1.ValkeyMaxMemoryPolicyAllkeysLRU,
		pgratorv1.ValkeyMaxMemoryPolicyAllkeysRandom,
		pgratorv1.ValkeyMaxMemoryPolicyNoEviction,
		pgratorv1.ValkeyMaxMemoryPolicyVolatileLFU,
		pgratorv1.ValkeyMaxMemoryPolicyVolatileLRU,
		pgratorv1.ValkeyMaxMemoryPolicyVolatileRandom,
		pgratorv1.ValkeyMaxMemoryPolicyVolatileTTL,
	}
	for _, p := range policies {
		t.Run(string(p), func(t *testing.T) {
			got, err := valkeyFromCRD(valkeyObj("v", pgratorv1.ValkeyTierSingleNode, pgratorv1.ValkeyMemory4GB, p))
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

// ---------------------------------------------------------------------------
// openSearchFromCRD
// ---------------------------------------------------------------------------

func TestOpenSearchFromCRD(t *testing.T) {
	for name, tc := range map[string]struct {
		obj     *pgratorv1.OpenSearch
		wantErr bool
	}{
		"single node v1": {
			obj: openSearchObj("idx", pgratorv1.OpenSearchTierSingleNode, pgratorv1.OpenSearchMemory2GB, pgratorv1.OpenSearchVersionV1, 10),
		},
		"high availability v2.19": {
			obj: openSearchObj("idx", pgratorv1.OpenSearchTierHighAvailability, pgratorv1.OpenSearchMemory32GB, pgratorv1.OpenSearchVersionV2_19, 200),
		},
		"bad tier": {
			obj:     openSearchObj("idx", "SuperNode", pgratorv1.OpenSearchMemory4GB, pgratorv1.OpenSearchVersionV2, 10),
			wantErr: true,
		},
		"bad memory": {
			obj:     openSearchObj("idx", pgratorv1.OpenSearchTierSingleNode, "3GB", pgratorv1.OpenSearchVersionV2, 10),
			wantErr: true,
		},
		"bad version": {
			obj:     openSearchObj("idx", pgratorv1.OpenSearchTierSingleNode, pgratorv1.OpenSearchMemory4GB, "99", 10),
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := openSearchFromCRD(tc.obj)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestOpenSearchFromCRD_AllMemorySizes(t *testing.T) {
	memories := []pgratorv1.OpenSearchMemory{
		pgratorv1.OpenSearchMemory2GB,
		pgratorv1.OpenSearchMemory4GB,
		pgratorv1.OpenSearchMemory8GB,
		pgratorv1.OpenSearchMemory16GB,
		pgratorv1.OpenSearchMemory32GB,
		pgratorv1.OpenSearchMemory64GB,
	}
	for _, mem := range memories {
		t.Run(string(mem), func(t *testing.T) {
			got, err := openSearchFromCRD(openSearchObj("idx", pgratorv1.OpenSearchTierSingleNode, mem, pgratorv1.OpenSearchVersionV2, 10))
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func TestOpenSearchFromCRD_AllVersions(t *testing.T) {
	versions := []pgratorv1.OpenSearchVersion{
		pgratorv1.OpenSearchVersionV1,
		pgratorv1.OpenSearchVersionV2,
		pgratorv1.OpenSearchVersionV2_19,
		pgratorv1.OpenSearchVersionV3_3,
	}
	for _, v := range versions {
		t.Run(string(v), func(t *testing.T) {
			got, err := openSearchFromCRD(openSearchObj("idx", pgratorv1.OpenSearchTierSingleNode, pgratorv1.OpenSearchMemory4GB, v, 10))
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func valkeyObj(name string, tier pgratorv1.ValkeyTier, mem pgratorv1.ValkeyMemory, policy pgratorv1.ValkeyMaxMemoryPolicy) *pgratorv1.Valkey {
	v := &pgratorv1.Valkey{}
	v.Name = name
	v.Spec.Tier = tier
	v.Spec.Memory = mem
	v.Spec.MaxMemoryPolicy = policy
	return v
}

func openSearchObj(name string, tier pgratorv1.OpenSearchTier, mem pgratorv1.OpenSearchMemory, version pgratorv1.OpenSearchVersion, storageGB int) *pgratorv1.OpenSearch {
	o := &pgratorv1.OpenSearch{}
	o.Name = name
	o.Spec.Tier = tier
	o.Spec.Memory = mem
	o.Spec.Version = version
	o.Spec.StorageGB = storageGB
	return o
}

// Ensure the package-level scheme has pgratorv1 types registered (compile-time check).
var (
	_ runtime.Object = (*pgratorv1.Valkey)(nil)
	_ runtime.Object = (*pgratorv1.OpenSearch)(nil)
)
