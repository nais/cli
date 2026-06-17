package apply

import (
	"reflect"
	"strings"
	"testing"
)

func TestDeepMerge(t *testing.T) {
	for name, tc := range map[string]struct {
		base     any
		override any
		want     any
	}{
		"override scalar wins": {
			base:     map[string]any{"a": 1},
			override: map[string]any{"a": 2},
			want:     map[string]any{"a": 2},
		},
		"new key added": {
			base:     map[string]any{"a": 1},
			override: map[string]any{"b": 2},
			want:     map[string]any{"a": 1, "b": 2},
		},
		"nested maps merge recursively": {
			base:     map[string]any{"spec": map[string]any{"image": "old", "replicas": 1}},
			override: map[string]any{"spec": map[string]any{"image": "new"}},
			want:     map[string]any{"spec": map[string]any{"image": "new", "replicas": 1}},
		},
		"lists are appended": {
			base:     map[string]any{"env": []any{"A"}},
			override: map[string]any{"env": []any{"B"}},
			want:     map[string]any{"env": []any{"A", "B"}},
		},
		"type change replaces": {
			base:     map[string]any{"a": map[string]any{"x": 1}},
			override: map[string]any{"a": "scalar"},
			want:     map[string]any{"a": "scalar"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := deepMerge(tc.base, tc.override)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("deepMerge() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestApplySet(t *testing.T) {
	t.Run("sets nested string", func(t *testing.T) {
		doc := map[string]any{"spec": map[string]any{}}
		if err := applySet(doc, "spec.image", "ghcr.io/nais/app:latest"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := doc["spec"].(map[string]any)["image"]; got != "ghcr.io/nais/app:latest" {
			t.Errorf("image = %v, want %v", got, "ghcr.io/nais/app:latest")
		}
	})

	t.Run("parses int", func(t *testing.T) {
		doc := map[string]any{}
		if err := applySet(doc, "spec.replicas", "3"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := doc["spec"].(map[string]any)["replicas"]; got != 3 {
			t.Errorf("replicas = %v (%T), want 3 (int)", got, got)
		}
	})

	t.Run("parses bool", func(t *testing.T) {
		doc := map[string]any{}
		if err := applySet(doc, "spec.enabled", "true"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := doc["spec"].(map[string]any)["enabled"]; got != true {
			t.Errorf("enabled = %v (%T), want true (bool)", got, got)
		}
	})

	t.Run("creates intermediate maps", func(t *testing.T) {
		doc := map[string]any{}
		if err := applySet(doc, "a.b.c", "v"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := doc["a"].(map[string]any)["b"].(map[string]any)["c"]; got != "v" {
			t.Errorf("a.b.c = %v, want v", got)
		}
	})

	t.Run("overrides existing value", func(t *testing.T) {
		doc := map[string]any{"spec": map[string]any{"image": "old"}}
		if err := applySet(doc, "spec.image", "new"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := doc["spec"].(map[string]any)["image"]; got != "new" {
			t.Errorf("image = %v, want new", got)
		}
	})

	t.Run("errors when intermediate is not a map", func(t *testing.T) {
		err := applySet(map[string]any{"spec": "scalar"}, "spec.image", "v")
		if err == nil || !strings.Contains(err.Error(), "is not a map") {
			t.Errorf("got %v, want error containing %q", err, "is not a map")
		}
	})

	t.Run("errors on empty key", func(t *testing.T) {
		if err := applySet(map[string]any{}, "", "v"); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("errors on empty path segment", func(t *testing.T) {
		if err := applySet(map[string]any{}, "spec..image", "v"); err == nil {
			t.Error("expected error, got nil")
		}
	})
}
