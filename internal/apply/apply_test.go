package apply

import (
	"testing"

	"github.com/nais/cli/internal/apply/native"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadManifestFile_Validation(t *testing.T) {
	_, err := readManifestFile("")
	assert.ErrorContains(t, err, "file path cannot be empty")

	_, err = readManifestFile("manifest.toml")
	assert.ErrorContains(t, err, "unsupported file extension")

	_, err = readManifestFile("does-not-exist.yaml")
	assert.ErrorContains(t, err, "failed to read file")
}

func TestToUnstructured(t *testing.T) {
	resources, err := native.Parse([]byte(`
version: v1
kind: SomeFutureKind
metadata:
  name: my-resource
spec:
  foo: bar
  nested:
    count: 3
`))
	require.NoError(t, err)
	require.Len(t, resources, 1)

	u, err := toUnstructured(resources[0])
	require.NoError(t, err)

	assert.Equal(t, "nais.io/v1", u.GetAPIVersion())
	assert.Equal(t, "SomeFutureKind", u.GetKind())
	assert.Equal(t, "my-resource", u.GetName())

	spec, found, err := unstructuredNestedMap(u.Object, "spec")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "bar", spec["foo"])
}

func TestDecodeCRD_PreservesFullManifest(t *testing.T) {
	docs, err := native.Documents([]byte(`
apiVersion: nais.io/v1alpha1
kind: Application
metadata:
  labels:
    team: examples
    label: value
  name: testapp
  namespace: examples
spec:
  image: testapp:1
  replicas:
    min: 1
    max: 1
`))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.False(t, native.IsNativeManifest(docs[0]))

	u, err := decodeCRD(docs[0])
	require.NoError(t, err)

	// apiVersion, namespace and labels must be preserved as-is.
	assert.Equal(t, "nais.io/v1alpha1", u.GetAPIVersion())
	assert.Equal(t, "Application", u.GetKind())
	assert.Equal(t, "testapp", u.GetName())
	assert.Equal(t, "examples", u.GetNamespace())
	assert.Equal(t, map[string]string{"team": "examples", "label": "value"}, u.GetLabels())
}

func TestDecodeCRD_MissingFields(t *testing.T) {
	docs, err := native.Documents([]byte("apiVersion: nais.io/v1alpha1\nmetadata:\n  name: x\n"))
	require.NoError(t, err)
	require.Len(t, docs, 1)

	_, err = decodeCRD(docs[0])
	assert.ErrorContains(t, err, `missing required field "kind"`)
}

// unstructuredNestedMap is a tiny helper to read a nested map without pulling in
// extra dependencies in the test.
func unstructuredNestedMap(obj map[string]any, key string) (map[string]any, bool, error) {
	v, ok := obj[key]
	if !ok {
		return nil, false, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false, assert.AnError
	}
	return m, true, nil
}
