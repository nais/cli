package resource

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nais/cli/internal/naisapi/gql"
	"gopkg.in/yaml.v3"
)

func TestParse_Valkey(t *testing.T) {
	manifests, err := Parse([]byte(`
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
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}

	m := manifests[0]
	if got, want := m.Version, "v1"; got != want {
		t.Errorf("version = %q, want %q", got, want)
	}
	if got, want := m.Kind, "Valkey"; got != want {
		t.Errorf("kind = %q, want %q", got, want)
	}
	if got, want := m.Name, "my-valkey-instance"; got != want {
		t.Errorf("name = %q, want %q", got, want)
	}
	if len(m.IgnoredFields) != 0 {
		t.Errorf("expected no ignored fields, got %v", m.IgnoredFields)
	}

	var spec valkeySpec
	if err := decodeSpec(&m.Spec, &spec); err != nil {
		t.Fatalf("decodeSpec: %v", err)
	}
	want := valkeySpec{
		Memory:               "4GB",
		Tier:                 "HighAvailability",
		Databases:            16,
		MaxMemoryPolicy:      "allkeys-lru",
		NotifyKeyspaceEvents: "Ex",
		Persistence:          &valkeyPersistence{Disabled: false},
	}
	if diff := cmp.Diff(want, spec); diff != "" {
		t.Errorf("spec mismatch (-want +got):\n%s", diff)
	}
}

func TestParse_IgnoredFields(t *testing.T) {
	manifests, err := Parse([]byte(`
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
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}

	want := []string{"apiVersion", "metadata.namespace", "metadata.annotations"}
	got := manifests[0].IgnoredFields
	if diff := cmp.Diff(want, got, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
		t.Errorf("ignored fields mismatch (-want +got):\n%s", diff)
	}
}

func TestParse_MultiDocument(t *testing.T) {
	manifests, err := Parse([]byte(`
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
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	want := []string{"Valkey", "OpenSearch"}
	got := []string{}
	for _, m := range manifests {
		got = append(got, m.Kind)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("kinds mismatch (-want +got):\n%s", diff)
	}
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
			mustErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestDecodeSpec_RejectsUnknownFields(t *testing.T) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte("memory: \"4GB\"\ntier: SingleNode\nbogus: true\n"), &node); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	var spec valkeySpec
	err := decodeSpec(node.Content[0], &spec)
	mustErrorContains(t, err, "bogus")
}

func TestEnumValue(t *testing.T) {
	got, err := enumValue("memory", "4GB", valkeyMemory)
	if err != nil {
		t.Fatalf("enumValue: %v", err)
	}
	if got != gql.ValkeyMemoryGb4 {
		t.Errorf("enumValue = %q, want %q", got, gql.ValkeyMemoryGb4)
	}

	_, err = enumValue("memory", "5GB", valkeyMemory)
	mustErrorContains(t, err, `invalid memory "5GB"`)
	// The error lists the allowed values.
	mustErrorContains(t, err, "4GB")
}

func TestIsNativeManifest(t *testing.T) {
	for name, tc := range map[string]struct {
		manifest string
		want     bool
	}{
		"native manifest": {
			manifest: "version: v1\nkind: Valkey\nmetadata:\n  name: x\nspec: {}\n",
			want:     true,
		},
		"regular CRD": {
			manifest: "apiVersion: nais.io/v1alpha1\nkind: Application\nmetadata:\n  name: testapp\n  namespace: examples\nspec: {}\n",
			want:     false,
		},
		// A document carrying apiVersion is never native, even with a version field.
		"both apiVersion and version": {
			manifest: "apiVersion: nais.io/v1alpha1\nversion: v1\nkind: Application\nmetadata:\n  name: testapp\nspec: {}\n",
			want:     false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := IsNativeManifest(mustDocument(t, tc.manifest)); got != tc.want {
				t.Errorf("IsNativeManifest = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDocuments_SplitsAndSkipsEmpty(t *testing.T) {
	docs, err := Documents([]byte("version: v1\nkind: Valkey\nmetadata:\n  name: a\nspec: {}\n---\n---\napiVersion: nais.io/v1alpha1\nkind: Application\nmetadata:\n  name: b\nspec: {}\n"))
	if err != nil {
		t.Fatalf("Documents: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}
	if !IsNativeManifest(docs[0]) {
		t.Error("expected docs[0] to be native")
	}
	if IsNativeManifest(docs[1]) {
		t.Error("expected docs[1] to be a regular CRD")
	}
}

func TestParse_Config(t *testing.T) {
	manifests, err := Parse([]byte(`
version: v1
kind: Config
metadata:
  name: my-config
  labels:
    purpose: backend
data:
  DATABASE_HOST: db.example.com
  LOG_LEVEL: info
  PORT: "8080"
binaryData:
  keystore.p12: aGVsbG8gd29ybGQ=
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}

	m := manifests[0]
	if got, want := m.Kind, "Config"; got != want {
		t.Errorf("kind = %q, want %q", got, want)
	}
	if got, want := m.Name, "my-config"; got != want {
		t.Errorf("name = %q, want %q", got, want)
	}
	if len(m.IgnoredFields) != 0 {
		t.Errorf("expected no ignored fields, got %v", m.IgnoredFields)
	}

	wantLabels := map[string]string{"purpose": "backend"}
	if diff := cmp.Diff(wantLabels, m.Labels); diff != "" {
		t.Errorf("labels mismatch (-want +got):\n%s", diff)
	}

	wantData := map[string]string{
		"DATABASE_HOST": "db.example.com",
		"LOG_LEVEL":     "info",
		"PORT":          "8080",
	}
	if diff := cmp.Diff(wantData, m.Data); diff != "" {
		t.Errorf("data mismatch (-want +got):\n%s", diff)
	}

	wantBinaryData := map[string]string{
		"keystore.p12": "aGVsbG8gd29ybGQ=",
	}
	if diff := cmp.Diff(wantBinaryData, m.BinaryData); diff != "" {
		t.Errorf("binaryData mismatch (-want +got):\n%s", diff)
	}
}

func TestParse_ConfigNoIgnoredFields(t *testing.T) {
	// Config uses data/binaryData at top level — these should NOT be ignored fields.
	manifests, err := Parse([]byte(`
version: v1
kind: Config
metadata:
  name: test
data:
  KEY: value
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(manifests[0].IgnoredFields) != 0 {
		t.Errorf("expected no ignored fields, got %v", manifests[0].IgnoredFields)
	}
}

func mustDocument(t *testing.T, manifest string) *yaml.Node {
	t.Helper()
	docs, err := Documents([]byte(manifest))
	if err != nil {
		t.Fatalf("Documents: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	return docs[0]
}

func mustErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not contain %q", err.Error(), want)
	}
}
