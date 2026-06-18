package labels

import "testing"

func TestParseAssignments(t *testing.T) {
	tests := []struct {
		name    string
		args    LabelFilters
		want    map[string]string
		wantErr bool
	}{
		{
			name: "single label",
			args: LabelFilters{"team=foo"},
			want: map[string]string{"team": "foo"},
		},
		{
			name: "multiple labels",
			args: LabelFilters{"team=foo", "domain=payments"},
			want: map[string]string{
				"team":   "foo",
				"domain": "payments",
			},
		},
		{
			name: "value may be empty",
			args: LabelFilters{"team="},
			want: map[string]string{"team": ""},
		},
		{
			name:    "missing equals",
			args:    LabelFilters{"team"},
			wantErr: true,
		},
		{
			name:    "missing key",
			args:    LabelFilters{"=foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAssignments(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d labels, want %d", len(got), len(tt.want))
			}
			for key, wantValue := range tt.want {
				if gotValue, ok := got[key]; !ok {
					t.Fatalf("missing key %q", key)
				} else if gotValue != wantValue {
					t.Fatalf("value for key %q = %q, want %q", key, gotValue, wantValue)
				}
			}
		})
	}
}
