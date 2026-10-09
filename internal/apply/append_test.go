package apply

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestApplyAppend(t *testing.T) {
	for _, tc := range []struct {
		name  string
		doc   map[string]any
		path  string
		value string
		want  map[string]any
	}{
		{
			name: "preserves existing elements",
			doc:  map[string]any{"list": []any{"first"}}, path: "list", value: "second",
			want: map[string]any{"list": []any{"first", "second"}},
		},
		{
			name: "appends map",
			doc:  map[string]any{"env": []any{map[string]any{"name": "EXISTING", "value": "keep"}}},
			path: "env", value: `{name: LOG_LEVEL, value: "hello, world=ok"}`,
			want: map[string]any{"env": []any{
				map[string]any{"name": "EXISTING", "value": "keep"},
				map[string]any{"name": "LOG_LEVEL", "value": "hello, world=ok"},
			}},
		},
		{
			name: "creates missing maps and list",
			doc:  map[string]any{}, path: "spec.env", value: "{name: LOG_LEVEL, value: debug}",
			want: map[string]any{"spec": map[string]any{"env": []any{map[string]any{"name": "LOG_LEVEL", "value": "debug"}}}},
		},
		{
			name: "creates missing list under existing map",
			doc:  map[string]any{"spec": map[string]any{"image": "keep"}}, path: "spec.args", value: "--debug",
			want: map[string]any{"spec": map[string]any{"image": "keep", "args": []any{"--debug"}}},
		},
		{
			name: "appends to empty list",
			doc:  map[string]any{"list": []any{}}, path: "list", value: "3",
			want: map[string]any{"list": []any{3}},
		},
		{
			name: "does not flatten YAML lists",
			doc:  map[string]any{"list": []any{"keep"}}, path: "list", value: "[one, two]",
			want: map[string]any{"list": []any{"keep", []any{"one", "two"}}},
		},
		{
			name: "appends to indexed nested list",
			doc:  map[string]any{"list": []any{[]any{"keep"}, []any{"other"}}}, path: "list[0]", value: "true",
			want: map[string]any{"list": []any{[]any{"keep", true}, []any{"other"}}},
		},
		{
			name: "creates list within indexed map",
			doc:  map[string]any{"groups": []any{map[string]any{"name": "keep"}}}, path: "groups[0].members", value: "alice",
			want: map[string]any{"groups": []any{map[string]any{"name": "keep", "members": []any{"alice"}}}},
		},
		{
			name: "allows duplicate elements",
			doc:  map[string]any{"list": []any{"same"}}, path: "list", value: "same",
			want: map[string]any{"list": []any{"same", "same"}},
		},
		{
			name: "appends null element",
			doc:  map[string]any{"list": []any{}}, path: "list", value: "null",
			want: map[string]any{"list": []any{nil}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := applyAppend(tc.doc, tc.path, tc.value); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.doc, tc.want) {
				t.Errorf("doc = %#v, want %#v", tc.doc, tc.want)
			}
		})
	}
}

func TestApplyAppendErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		doc   map[string]any
		path  string
		value string
		want  string
	}{
		{"scalar target", map[string]any{"list": "keep"}, "list", "new", "target is not a list"},
		{"map target", map[string]any{"list": map[string]any{}}, "list", "new", "target is not a list"},
		{"null target", map[string]any{"list": nil}, "list", "new", "target is not a list"},
		{"scalar parent", map[string]any{"spec": "keep"}, "spec.list", "new", "not a map"},
		{"null parent", map[string]any{"spec": nil}, "spec.list", "new", "not a map"},
		{"missing indexed list", map[string]any{}, "spec.groups[0].members", "new", "existing list"},
		{"out of range", map[string]any{"list": []any{}}, "list[0]", "new", "out of range"},
		{"invalid path", map[string]any{}, "list[-1]", "new", "invalid --append key"},
		{"empty path", map[string]any{}, "", "new", "empty --append key"},
		{"invalid YAML", map[string]any{}, "spec.list", "[", "invalid --append value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, err := yaml.Marshal(tc.doc)
			if err != nil {
				t.Fatal(err)
			}
			err = applyAppend(tc.doc, tc.path, tc.value)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
			after, err := yaml.Marshal(tc.doc)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Errorf("document mutated on error: %s", after)
			}
		})
	}
}
