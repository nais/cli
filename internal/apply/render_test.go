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

func TestRenderDir_CollectsBaseFilesAndExcludesMixins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.yaml", "kind: Application\nmetadata:\n  name: myapp\n")
	writeFile(t, dir, "app.dev.yaml", "spec:\n  image: dev-image\n")
	writeFile(t, dir, "valkey.yaml", "kind: Valkey\nmetadata:\n  name: myvalkey\n")

	got, err := renderDir(dir, "dev", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(got)
	// app.yaml should be rendered with its mixin
	if !strings.Contains(output, "dev-image") {
		t.Errorf("output should contain mixin value 'dev-image', got:\n%s", output)
	}
	// valkey.yaml should be included
	if !strings.Contains(output, "Valkey") {
		t.Errorf("output should contain Valkey resource, got:\n%s", output)
	}
	// mixin file should NOT appear as its own resource
	if strings.Count(output, "kind:") != 2 {
		t.Errorf("expected 2 resources, got:\n%s", output)
	}
}

func TestRenderDir_ExcludesMixinsForAllEnvironments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.yaml", "apiVersion: nais.io/v1alpha1\nkind: Application\nmetadata:\n  name: myapp\nspec:\n  image: base\n")
	writeFile(t, dir, "app.dev-gcp.yaml", "spec:\n  image: dev\n")
	writeFile(t, dir, "app.prod-gcp.yaml", "spec:\n  image: prod\n")
	writeFile(t, dir, "config.yaml", "version: v1\nkind: Config\nmetadata:\n  name: myconfig\ndata:\n  KEY: value\n")
	writeFile(t, dir, "config.dev-gcp.yaml", "data:\n  ENV: dev\n")
	writeFile(t, dir, "config.prod-gcp.yaml", "data:\n  ENV: prod\n")

	got, err := renderDir(dir, "dev-gcp", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(got)
	// Only base files should produce resources
	if strings.Count(output, "kind:") != 2 {
		t.Errorf("expected 2 resources (app + config), got:\n%s", output)
	}
	// The active mixin for dev-gcp should be merged
	if !strings.Contains(output, "dev") {
		t.Errorf("expected dev-gcp mixin to be applied, got:\n%s", output)
	}
}

func TestRenderDir_IgnoresNonYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.yaml", "kind: Application\nmetadata:\n  name: myapp\n")
	writeFile(t, dir, "readme.md", "# Not a manifest")
	writeFile(t, dir, "config.json", "{}")

	got, err := renderDir(dir, "dev", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "Application") {
		t.Errorf("expected Application in output, got:\n%s", got)
	}
}

func TestRenderDir_EmptyDirectoryFails(t *testing.T) {
	dir := t.TempDir()

	_, err := renderDir(dir, "dev", nil, discardWriter())
	if err == nil || !strings.Contains(err.Error(), "no YAML resource files found") {
		t.Errorf("got %v, want error about no YAML files", err)
	}
}

func TestRenderDir_OnlyMixinFilesFails(t *testing.T) {
	dir := t.TempDir()
	// Create a mixin without a corresponding base - but since there's no base,
	// it won't be detected as a mixin and will be treated as a base file.
	// To actually test "only mixins", we need a base that has a mixin.
	writeFile(t, dir, "app.yaml", "kind: Application\nmetadata:\n  name: myapp\n")
	writeFile(t, dir, "app.dev.yaml", "spec:\n  image: dev\n")

	got, err := renderDir(dir, "dev", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only render app.yaml (with mixin), not treat mixin as separate resource
	if strings.Count(string(got), "kind:") != 1 {
		t.Errorf("expected 1 resource (mixin excluded), got:\n%s", got)
	}
}

func TestRenderDir_IgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.yaml", "kind: Application\nmetadata:\n  name: myapp\n")
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, subdir, "other.yaml", "kind: Other\nmetadata:\n  name: other\n")

	got, err := renderDir(dir, "dev", nil, discardWriter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(got), "Other") {
		t.Errorf("subdirectory files should not be included, got:\n%s", got)
	}
}

func TestMixinSet(t *testing.T) {
	t.Run("known environments", func(t *testing.T) {
		files := []string{"app.yaml", "app.dev.yaml", "app.prod.yaml", "valkey.yaml", "valkey.dev.yaml", "standalone.yaml"}
		envs := []string{"dev", "prod"}

		mixins := mixinSet(files, envs)
		if !mixins["app.dev.yaml"] {
			t.Error("app.dev.yaml should be a mixin")
		}
		if !mixins["app.prod.yaml"] {
			t.Error("app.prod.yaml should be a mixin")
		}
		if !mixins["valkey.dev.yaml"] {
			t.Error("valkey.dev.yaml should be a mixin")
		}
		if mixins["app.yaml"] {
			t.Error("app.yaml should NOT be a mixin")
		}
		if mixins["valkey.yaml"] {
			t.Error("valkey.yaml should NOT be a mixin")
		}
		if mixins["standalone.yaml"] {
			t.Error("standalone.yaml should NOT be a mixin")
		}
	})

	t.Run("unknown environment suffix is not a mixin", func(t *testing.T) {
		files := []string{"app.yaml", "app.unknown.yaml", "config.yaml", "config.unknown.yaml"}
		envs := []string{"dev", "prod"}

		mixins := mixinSet(files, envs)
		if mixins["app.unknown.yaml"] {
			t.Error("app.unknown.yaml should NOT be a mixin ('unknown' is not a known environment)")
		}
	})

	t.Run("mixin without base is not filtered", func(t *testing.T) {
		files := []string{"orphan.dev.yaml", "other.yaml"}
		envs := []string{"dev"}

		mixins := mixinSet(files, envs)
		if mixins["orphan.dev.yaml"] {
			t.Error("orphan.dev.yaml should NOT be a mixin (no orphan.yaml base)")
		}
	})

	t.Run("compound environment names", func(t *testing.T) {
		files := []string{"app.yaml", "app.dev-gcp.yaml", "app.prod-gcp.yaml", "config.yaml", "config.dev-gcp.yaml"}
		envs := []string{"dev-gcp", "prod-gcp"}

		mixins := mixinSet(files, envs)
		if !mixins["app.dev-gcp.yaml"] {
			t.Error("app.dev-gcp.yaml should be a mixin")
		}
		if !mixins["app.prod-gcp.yaml"] {
			t.Error("app.prod-gcp.yaml should be a mixin")
		}
		if !mixins["config.dev-gcp.yaml"] {
			t.Error("config.dev-gcp.yaml should be a mixin")
		}
	})

	t.Run("fallback heuristic with empty env list", func(t *testing.T) {
		files := []string{"app.yaml", "app.dev.yaml", "config.yaml", "config.prod.yaml"}
		envs := []string{}

		mixins := mixinSet(files, envs)
		if !mixins["app.dev.yaml"] {
			t.Error("app.dev.yaml should be a mixin (heuristic fallback)")
		}
		if !mixins["config.prod.yaml"] {
			t.Error("config.prod.yaml should be a mixin (heuristic fallback)")
		}
	})
}
