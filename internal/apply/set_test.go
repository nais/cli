package apply

import (
	"reflect"
	"strings"
	"testing"
)

func TestApplySetLists(t *testing.T) {
	for _, tc := range []struct {
		name  string
		doc   map[string]any
		path  string
		value string
		want  map[string]any
	}{
		{
			name: "replaces entire list",
			doc: map[string]any{"env": []any{
				map[string]any{"name": "OLD", "value": "old"},
			}},
			path: "env", value: "[{name: LOG_LEVEL, value: debug}]",
			want: map[string]any{"env": []any{
				map[string]any{"name": "LOG_LEVEL", "value": "debug"},
			}},
		},
		{
			name: "creates list through missing maps",
			doc:  map[string]any{}, path: "spec.ingresses",
			value: `["https://a.example.com", "https://b.example.com"]`,
			want:  map[string]any{"spec": map[string]any{"ingresses": []any{"https://a.example.com", "https://b.example.com"}}},
		},
		{
			name: "clears list",
			doc:  map[string]any{"list": []any{"old"}}, path: "list", value: "[]",
			want: map[string]any{"list": []any{}},
		},
		{
			name: "replaces indexed scalar",
			doc:  map[string]any{"list": []any{"first", "second"}}, path: "list[1]", value: "new",
			want: map[string]any{"list": []any{"first", "new"}},
		},
		{
			name: "updates field in indexed map",
			doc: map[string]any{"env": []any{
				map[string]any{"name": "LOG_LEVEL", "value": "info"},
				map[string]any{"name": "OTHER", "value": "keep"},
			}},
			path: "env[0].value", value: "debug",
			want: map[string]any{"env": []any{
				map[string]any{"name": "LOG_LEVEL", "value": "debug"},
				map[string]any{"name": "OTHER", "value": "keep"},
			}},
		},
		{
			name: "creates missing maps within indexed map",
			doc:  map[string]any{"list": []any{map[string]any{}}}, path: "list[0].nested.value", value: "3",
			want: map[string]any{"list": []any{map[string]any{"nested": map[string]any{"value": 3}}}},
		},
		{
			name: "nested lists",
			doc:  map[string]any{"list": []any{[]any{"old", "keep"}}}, path: "list[0][0]", value: "true",
			want: map[string]any{"list": []any{[]any{true, "keep"}}},
		},
		{
			name: "multiple lists in path",
			doc:  map[string]any{"list": []any{map[string]any{"items": []any{"old"}}}},
			path: "list[0].items[0]", value: "{name: new}",
			want: map[string]any{"list": []any{map[string]any{"items": []any{map[string]any{"name": "new"}}}}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := applySet(tc.doc, tc.path, tc.value); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.doc, tc.want) {
				t.Errorf("doc = %#v, want %#v", tc.doc, tc.want)
			}
		})
	}
}

func TestApplySetInvalidListPaths(t *testing.T) {
	for _, path := range []string{
		"list[]", "list[-1]", "list[+1]", "list[abc]", "list[1.0]",
		"list[0", "list0]", "list[0]value", "list[0]]", "[0]",
		"list.[0]", "list[0].", "list[0]..value", "list[0][",
		"list[9999999999999999999999999999999999]",
	} {
		t.Run(path, func(t *testing.T) {
			doc := map[string]any{"list": []any{"old"}}
			err := applySet(doc, path, "new")
			if err == nil || !strings.Contains(err.Error(), "invalid --set key") {
				t.Fatalf("got %v, want invalid key error", err)
			}
			if !reflect.DeepEqual(doc, map[string]any{"list": []any{"old"}}) {
				t.Errorf("document mutated on error: %#v", doc)
			}
		})
	}
}

func TestApplySetListTraversalErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  map[string]any
		path string
		want string
	}{
		{"missing list", map[string]any{}, "spec.env[0].value", "existing list"},
		{"scalar instead of list", map[string]any{"list": "old"}, "list[0]", "existing list"},
		{"map instead of list", map[string]any{"list": map[string]any{}}, "list[0]", "existing list"},
		{"null instead of list", map[string]any{"list": nil}, "list[0]", "existing list"},
		{"empty list", map[string]any{"list": []any{}}, "list[0]", "out of range"},
		{"index at length", map[string]any{"list": []any{"old"}}, "list[1]", "out of range"},
		{"index beyond length", map[string]any{"list": []any{"old"}}, "list[5]", "out of range"},
		{"scalar element", map[string]any{"list": []any{"old"}}, "list[0].value", "not a map"},
		{"null element", map[string]any{"list": []any{nil}}, "list[0].value", "not a map"},
		{"unindexed list", map[string]any{"list": []any{"old"}}, "list.value", "not a map"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := applySet(tc.doc, tc.path, "new")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
		})
	}
}
