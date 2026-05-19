package app

import (
	"testing"
)

func TestParseEnvVarUpdates(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []EnvVarUpdate
		wantErr bool
	}{
		{
			name: "set single variable",
			args: []string{"FOO=bar"},
			want: []EnvVarUpdate{{Name: "FOO", Value: ptr("bar")}},
		},
		{
			name: "set multiple variables",
			args: []string{"FOO=bar", "BAZ=qux"},
			want: []EnvVarUpdate{
				{Name: "FOO", Value: ptr("bar")},
				{Name: "BAZ", Value: ptr("qux")},
			},
		},
		{
			name: "remove variable",
			args: []string{"FOO-"},
			want: []EnvVarUpdate{{Name: "FOO", Value: nil}},
		},
		{
			name: "set and remove",
			args: []string{"NEW=hello", "OLD-"},
			want: []EnvVarUpdate{
				{Name: "NEW", Value: ptr("hello")},
				{Name: "OLD", Value: nil},
			},
		},
		{
			name: "value with equals sign",
			args: []string{"FOO=bar=baz"},
			want: []EnvVarUpdate{{Name: "FOO", Value: ptr("bar=baz")}},
		},
		{
			name: "empty value",
			args: []string{"FOO="},
			want: []EnvVarUpdate{{Name: "FOO", Value: ptr("")}},
		},
		{
			name: "value with hyphen",
			args: []string{"FOO=bar-baz"},
			want: []EnvVarUpdate{{Name: "FOO", Value: ptr("bar-baz")}},
		},
		{
			name: "key with hyphen and delete",
			args: []string{"MY-VAR-"},
			want: []EnvVarUpdate{{Name: "MY-VAR", Value: nil}},
		},
		{
			name:    "empty key on set",
			args:    []string{"=value"},
			wantErr: true,
		},
		{
			name:    "just a hyphen",
			args:    []string{"-"},
			wantErr: true,
		},
		{
			name:    "no equals and no trailing hyphen",
			args:    []string{"FOO"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnvVarUpdates(tt.args)
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
				t.Fatalf("got %v updates, want %v", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("update[%v].Name = %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Value == nil && tt.want[i].Value == nil {
					continue
				}
				if got[i].Value == nil || tt.want[i].Value == nil {
					t.Errorf("update[%v].Value pointer mismatch: got %v, want %v", i, got[i].Value, tt.want[i].Value)
					continue
				}
				if *got[i].Value != *tt.want[i].Value {
					t.Errorf("update[%v].Value = %q, want %q", i, *got[i].Value, *tt.want[i].Value)
				}
			}
		})
	}
}

func ptr(s string) *string {
	return &s
}
