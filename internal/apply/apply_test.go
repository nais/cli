package apply

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	applyflag "github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/apply/resource"
	flagspkg "github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
	"gopkg.in/yaml.v3"
)

func TestReadManifestFile_Validation(t *testing.T) {
	for name, tc := range map[string]struct {
		path   string
		errMsg string
	}{
		"empty path":       {path: "", errMsg: "file path cannot be empty"},
		"bad extension":    {path: "manifest.toml", errMsg: "unsupported file extension"},
		"nonexistent file": {path: "does-not-exist.yaml", errMsg: "failed to read file"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := readManifestFile(tc.path)
			mustErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestToUnstructured(t *testing.T) {
	manifests, err := resource.Parse([]byte(`
version: v1
kind: SomeFutureKind
metadata:
  name: my-resource
spec:
  foo: bar
  nested:
    count: 3
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}

	u, err := toUnstructured(manifests[0], nil)
	if err != nil {
		t.Fatalf("toUnstructured: %v", err)
	}

	want := map[string]any{
		"apiVersion": "nais.io/v1",
		"kind":       "SomeFutureKind",
		"metadata":   map[string]any{"name": "my-resource"},
		"spec": map[string]any{
			"foo":    "bar",
			"nested": map[string]any{"count": 3},
		},
	}
	if diff := cmp.Diff(want, u.Object); diff != "" {
		t.Errorf("unstructured mismatch (-want +got):\n%s", diff)
	}
}

func TestDecodeCRD_PreservesFullManifest(t *testing.T) {
	docs, err := resource.Documents([]byte(`
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
	if err != nil {
		t.Fatalf("Documents: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if resource.IsNativeManifest(docs[0]) {
		t.Fatal("expected a regular CRD, got a native manifest")
	}

	u, err := decodeCRD(docs[0])
	if err != nil {
		t.Fatalf("decodeCRD: %v", err)
	}

	// apiVersion, namespace and labels must be preserved as-is.
	if got, want := u.GetAPIVersion(), "nais.io/v1alpha1"; got != want {
		t.Errorf("apiVersion = %q, want %q", got, want)
	}
	if got, want := u.GetKind(), "Application"; got != want {
		t.Errorf("kind = %q, want %q", got, want)
	}
	if got, want := u.GetName(), "testapp"; got != want {
		t.Errorf("name = %q, want %q", got, want)
	}
	if got, want := u.GetNamespace(), "examples"; got != want {
		t.Errorf("namespace = %q, want %q", got, want)
	}
	want := map[string]string{"team": "examples", "label": "value"}
	if diff := cmp.Diff(want, u.GetLabels()); diff != "" {
		t.Errorf("labels mismatch (-want +got):\n%s", diff)
	}
}

func TestDecodeCRD_MissingFields(t *testing.T) {
	docs, err := resource.Documents([]byte("apiVersion: nais.io/v1alpha1\nmetadata:\n  name: x\n"))
	if err != nil {
		t.Fatalf("Documents: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}

	_, err = decodeCRD(docs[0])
	mustErrorContains(t, err, `missing required field "kind"`)
}

func TestRun_DryRunDoesNotApply(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "nais.yaml")
	manifest := `
version: v1
kind: Application
metadata:
  name: myapp
spec:
  image: ghcr.io/nais/app:latest
---
apiVersion: nais.io/v1
kind: SomeCRD
metadata:
  name: custom
spec:
  foo: bar
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var out bytes.Buffer
	flags := &applyflag.Apply{
		GlobalFlags: &flagspkg.GlobalFlags{
			AdditionalFlags: &flagspkg.AdditionalFlags{
				Team:        "my-team",
				Environment: "dev",
			},
		},
		DryRun: true,
	}

	err := Run(context.Background(), manifestPath, flags, naistrix.NewOutputWriter(&out, new(naistrix.Count)))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Application/myapp: would apply",
		"SomeCRD/custom: would apply",
		"dry-run complete: no resources were applied",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}

func TestRun_DryRunFailsOnIgnoredFieldsWithoutAllowFlag(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "nais.yaml")
	manifest := `
version: v1
kind: Application
metadata:
  name: myapp
  namespace: should-fail
spec:
  image: ghcr.io/nais/app:latest
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	flags := &applyflag.Apply{
		GlobalFlags: &flagspkg.GlobalFlags{
			AdditionalFlags: &flagspkg.AdditionalFlags{
				Team:        "my-team",
				Environment: "dev",
			},
		},
		DryRun: true,
	}

	err := Run(context.Background(), manifestPath, flags, naistrix.NewOutputWriter(io.Discard, new(naistrix.Count)))
	mustErrorContains(t, err, "contains fields not used by nais apply")
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

func TestSortDocsWorkloadsLast(t *testing.T) {
	mkDoc := func(kind string) *yaml.Node {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "kind"},
				{Kind: yaml.ScalarNode, Value: kind},
			},
		}
	}

	docs := []*yaml.Node{
		mkDoc("Application"),
		mkDoc("Valkey"),
		mkDoc("Naisjob"),
		mkDoc("OpenSearch"),
		mkDoc("Config"),
	}

	sortDocsWorkloadsLast(docs)

	kinds := make([]string, len(docs))
	for i, d := range docs {
		kinds[i] = docKind(d)
	}

	// Workloads should be at the end
	for i, kind := range kinds {
		if workloadKinds[kind] && i < 3 {
			t.Errorf("workload %q at index %d, expected at index >= 3", kind, i)
		}
		if !workloadKinds[kind] && i >= 3 {
			t.Errorf("non-workload %q at index %d, expected at index < 3", kind, i)
		}
	}

	// Non-workloads preserve relative order
	nonWorkloads := []string{}
	for _, k := range kinds {
		if !workloadKinds[k] {
			nonWorkloads = append(nonWorkloads, k)
		}
	}
	if want := []string{"Valkey", "OpenSearch", "Config"}; strings.Join(nonWorkloads, ",") != strings.Join(want, ",") {
		t.Errorf("non-workload order = %v, want %v", nonWorkloads, want)
	}

	// Workloads preserve relative order
	workloads := []string{}
	for _, k := range kinds {
		if workloadKinds[k] {
			workloads = append(workloads, k)
		}
	}
	if want := []string{"Application", "Naisjob"}; strings.Join(workloads, ",") != strings.Join(want, ",") {
		t.Errorf("workload order = %v, want %v", workloads, want)
	}
}

func TestRun_DirectoryWithSetFails(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifests")
	if err := os.Mkdir(manifestPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestPath, "app.yaml"), []byte("kind: Application\nmetadata:\n  name: myapp\nspec:\n  image: test\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	flags := &applyflag.Apply{
		GlobalFlags: &flagspkg.GlobalFlags{
			AdditionalFlags: &flagspkg.AdditionalFlags{
				Team:        "my-team",
				Environment: "dev",
			},
		},
		Set: []string{"spec.image=override"},
	}

	err := Run(context.Background(), manifestPath, flags, naistrix.NewOutputWriter(io.Discard, new(naistrix.Count)))
	mustErrorContains(t, err, "--set cannot be used when applying a directory")
}

func TestRun_DirectoryWithMixinFlagFails(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifests")
	if err := os.Mkdir(manifestPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestPath, "app.yaml"), []byte("kind: Application\nmetadata:\n  name: myapp\nspec:\n  image: test\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	flags := &applyflag.Apply{
		GlobalFlags: &flagspkg.GlobalFlags{
			AdditionalFlags: &flagspkg.AdditionalFlags{
				Team:        "my-team",
				Environment: "dev",
			},
		},
		Mixin: "some-mixin.yaml",
	}

	err := Run(context.Background(), manifestPath, flags, naistrix.NewOutputWriter(io.Discard, new(naistrix.Count)))
	mustErrorContains(t, err, "--mixin cannot be used when applying a directory")
}
