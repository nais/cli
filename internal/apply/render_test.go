package apply

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/naistrix"
	"gopkg.in/yaml.v3"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("writing %s: %v", p, err)
	}
	return p
}

func renderToMap(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshalling rendered output: %v", err)
	}
	return m
}

func discardWriter() *naistrix.OutputWriter {
	return naistrix.NewOutputWriter(io.Discard, new(naistrix.Count))
}

func TestRender_NoOverridesReturnsBaseUnchanged(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "---\nkind: A\n---\nkind: B\n")

	got, err := render(base, "", "", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "---\nkind: A\n---\nkind: B\n"; string(got) != want {
		t.Errorf("render() = %q, want %q", got, want)
	}
}

func TestRender_ExplicitMixinDeepMerges(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "kind: Application\nspec:\n  image: old\n  replicas: 1\n")
	mixin := writeFile(t, dir, "dev.yaml", "spec:\n  image: new\n")

	got, err := render(base, mixin, "", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spec := renderToMap(t, got)["spec"].(map[string]any)
	if spec["image"] != "new" {
		t.Errorf("image = %v, want new", spec["image"])
	}
	if spec["replicas"] != 1 {
		t.Errorf("replicas = %v, want 1", spec["replicas"])
	}
}

func TestRender_AutoLoadsEnvMixin(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "kind: Application\nspec:\n  image: old\n")
	writeFile(t, dir, "nais.dev.yaml", "spec:\n  image: dev\n")

	got, err := render(base, "", "dev", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if image := renderToMap(t, got)["spec"].(map[string]any)["image"]; image != "dev" {
		t.Errorf("image = %v, want dev", image)
	}
}

func TestRender_SetWinsOverMixinAndBase(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "kind: Application\nspec:\n  image: base\n")
	mixin := writeFile(t, dir, "dev.yaml", "spec:\n  image: mixin\n")

	got, err := render(base, mixin, "", []string{"spec.image=set"}, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if image := renderToMap(t, got)["spec"].(map[string]any)["image"]; image != "set" {
		t.Errorf("image = %v, want set", image)
	}
}

func TestRender_SetParsesYAMLValue(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "kind: Application\nspec: {}\n")

	got, err := render(base, "", "", []string{"spec.replicas=3"}, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replicas := renderToMap(t, got)["spec"].(map[string]any)["replicas"]; replicas != 3 {
		t.Errorf("replicas = %v (%T), want 3 (int)", replicas, replicas)
	}
}

func TestRender_MultiDocWithMixinFails(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "kind: A\n---\nkind: B\n")
	mixin := writeFile(t, dir, "dev.yaml", "spec: {}\n")

	_, err := render(base, mixin, "", nil, discardWriter())
	if err == nil || !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Errorf("got %v, want error containing %q", err, "multiple YAML documents")
	}
}

func TestRender_InvalidSetFails(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.yaml", "kind: A\n")

	_, err := render(base, "", "", []string{"noequals"}, discardWriter())
	if err == nil || !strings.Contains(err.Error(), "expected KEY=VALUE") {
		t.Errorf("got %v, want error containing %q", err, "expected KEY=VALUE")
	}
}

func TestRender_UnsupportedExtensionFails(t *testing.T) {
	dir := t.TempDir()
	base := writeFile(t, dir, "nais.json", "{}")

	_, err := render(base, "", "", nil, discardWriter())
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("got %v, want error containing %q", err, "unsupported file extension")
	}
}
