package native

import (
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParse_Valkey(t *testing.T) {
	manifest := []byte(`
version: v1
kind: Valkey
metadata:
  name: my-valkey-instance
spec:
  memory: "4GB"
  tier: HighAvailability
  databases: 16
  maxMemoryPolicy: allkeys-lru
  notifyKeyspaceEvents: "Ex"
  persistence:
    disabled: false
`)

	resources, err := Parse(manifest)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	r := resources[0]
	assert.Equal(t, "v1", r.Version)
	assert.Equal(t, "Valkey", r.Kind)
	assert.Equal(t, "my-valkey-instance", r.Name)
	assert.Empty(t, r.IgnoredFields)

	var spec valkeySpec
	require.NoError(t, decodeSpec(&r.Spec, &spec))
	assert.Equal(t, "4GB", spec.Memory)
	assert.Equal(t, "HighAvailability", spec.Tier)
	assert.Equal(t, 16, spec.Databases)
	assert.Equal(t, "allkeys-lru", spec.MaxMemoryPolicy)
	assert.Equal(t, "Ex", spec.NotifyKeyspaceEvents)
	require.NotNil(t, spec.Persistence)
	assert.False(t, spec.Persistence.Disabled)
}

func TestParse_IgnoredFields(t *testing.T) {
	manifest := []byte(`
apiVersion: nais.io/v1
version: v1
kind: Valkey
metadata:
  name: my-valkey-instance
  namespace: default
  annotations:
    foo: bar
spec:
  memory: "4GB"
  tier: HighAvailability
`)

	resources, err := Parse(manifest)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	assert.ElementsMatch(
		t,
		[]string{"apiVersion", "metadata.namespace", "metadata.annotations"},
		resources[0].IgnoredFields,
	)
}

func TestParse_MultiDocument(t *testing.T) {
	manifest := []byte(`
version: v1
kind: Valkey
metadata:
  name: cache
spec:
  memory: "1GB"
  tier: SingleNode
---
version: v1
kind: OpenSearch
metadata:
  name: search
spec:
  memory: "8GB"
  tier: HighAvailability
  version: "2.19"
  storageGB: 100
`)

	resources, err := Parse(manifest)
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, "Valkey", resources[0].Kind)
	assert.Equal(t, "OpenSearch", resources[1].Kind)
}

func TestParse_Errors(t *testing.T) {
	for name, tc := range map[string]struct {
		manifest string
		errMsg   string
	}{
		"missing kind": {
			manifest: "version: v1\nmetadata:\n  name: x\nspec: {}\n",
			errMsg:   `missing required field "kind"`,
		},
		"missing name": {
			manifest: "version: v1\nkind: Valkey\nspec: {}\n",
			errMsg:   `missing required field "metadata.name"`,
		},
		"missing version": {
			manifest: "kind: Valkey\nmetadata:\n  name: x\nspec: {}\n",
			errMsg:   `missing required field "version"`,
		},
		"unsupported version": {
			manifest: "version: v2\nkind: Valkey\nmetadata:\n  name: x\nspec: {}\n",
			errMsg:   `unsupported version "v2"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte(tc.manifest))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestDecodeSpec_RejectsUnknownFields(t *testing.T) {
	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("memory: \"4GB\"\ntier: SingleNode\nbogus: true\n"), &node))

	var spec valkeySpec
	err := decodeSpec(node.Content[0], &spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}

func TestEnumValue(t *testing.T) {
	got, err := enumValue("memory", "4GB", valkeyMemory)
	require.NoError(t, err)
	assert.Equal(t, gql.ValkeyMemoryGb4, got)

	_, err = enumValue("memory", "5GB", valkeyMemory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid memory "5GB"`)
	// error lists allowed values
	assert.Contains(t, err.Error(), "4GB")
}

func TestHasMutation(t *testing.T) {
	assert.True(t, HasMutation("Valkey"))
	assert.True(t, HasMutation("OpenSearch"))
	assert.False(t, HasMutation("Unknown"))
}

func TestIsNativeManifest(t *testing.T) {
	nativeDoc := mustDocument(t, `
version: v1
kind: Valkey
metadata:
  name: x
spec: {}
`)
	assert.True(t, IsNativeManifest(nativeDoc))

	crdDoc := mustDocument(t, `
apiVersion: nais.io/v1alpha1
kind: Application
metadata:
  name: testapp
  namespace: examples
spec: {}
`)
	assert.False(t, IsNativeManifest(crdDoc))

	// A document carrying apiVersion is never native, even if it also has a
	// version field.
	bothDoc := mustDocument(t, `
apiVersion: nais.io/v1alpha1
version: v1
kind: Application
metadata:
  name: testapp
spec: {}
`)
	assert.False(t, IsNativeManifest(bothDoc))
}

func TestDocuments_SplitsAndSkipsEmpty(t *testing.T) {
	docs, err := Documents([]byte("version: v1\nkind: Valkey\nmetadata:\n  name: a\nspec: {}\n---\n---\napiVersion: nais.io/v1alpha1\nkind: Application\nmetadata:\n  name: b\nspec: {}\n"))
	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.True(t, IsNativeManifest(docs[0]))
	assert.False(t, IsNativeManifest(docs[1]))
}

func mustDocument(t *testing.T, manifest string) *yaml.Node {
	t.Helper()
	docs, err := Documents([]byte(manifest))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	return docs[0]
}
