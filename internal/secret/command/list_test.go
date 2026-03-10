package command

import "testing"

func TestSummarizeList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{
			name:  "empty",
			items: nil,
			want:  "",
		},
		{
			name:  "single item",
			items: []string{"a"},
			want:  "a",
		},
		{
			name:  "two items",
			items: []string{"a", "b"},
			want:  "a, b",
		},
		{
			name:  "exactly max items",
			items: []string{"a", "b", "c"},
			want:  "a, b, c",
		},
		{
			name:  "one over max",
			items: []string{"a", "b", "c", "d"},
			want:  "a, b, c, +1 more",
		},
		{
			name:  "many over max",
			items: []string{"a", "b", "c", "d", "e", "f"},
			want:  "a, b, c, +3 more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := summarizeList(tt.items)
			if got != tt.want {
				t.Errorf("summarizeList(%v) = %q, want %q", tt.items, got, tt.want)
			}
		})
	}
}
