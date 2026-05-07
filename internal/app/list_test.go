package app

import (
	"reflect"
	"testing"
)

func TestApplicationNames_Unique(t *testing.T) {
	tests := []struct {
		name  string
		input ApplicationNames
		want  []string
	}{
		{
			name:  "nil map",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty map",
			input: ApplicationNames{},
			want:  nil,
		},
		{
			name: "single environment",
			input: ApplicationNames{
				"dev": {"foo", "bar"},
			},
			want: []string{"bar", "foo"},
		},
		{
			name: "deduplicates across environments",
			input: ApplicationNames{
				"dev":  {"foo", "bar"},
				"prod": {"bar", "baz"},
			},
			want: []string{"bar", "baz", "foo"},
		},
		{
			name: "environment with no applications",
			input: ApplicationNames{
				"dev":  {"foo"},
				"prod": {},
			},
			want: []string{"foo"},
		},
		{
			name: "duplicates within same environment",
			input: ApplicationNames{
				"dev": {"foo", "foo", "bar"},
			},
			want: []string{"bar", "foo"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.input.Unique()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Unique() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestApplicationNames_InEnv(t *testing.T) {
	tests := []struct {
		name  string
		input ApplicationNames
		env   string
		want  []string
	}{
		{
			name:  "nil map",
			input: nil,
			env:   "dev",
			want:  nil,
		},
		{
			name: "missing environment",
			input: ApplicationNames{
				"dev": {"foo"},
			},
			env:  "prod",
			want: nil,
		},
		{
			name: "returns sorted apps for environment",
			input: ApplicationNames{
				"dev":  {"charlie", "alpha", "bravo"},
				"prod": {"zeta"},
			},
			env:  "dev",
			want: []string{"alpha", "bravo", "charlie"},
		},
		{
			name: "empty environment slice",
			input: ApplicationNames{
				"dev": {},
			},
			env:  "dev",
			want: []string{},
		},
		{
			name: "isolates environments",
			input: ApplicationNames{
				"dev":  {"foo"},
				"prod": {"bar"},
			},
			env:  "prod",
			want: []string{"bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.input.InEnv(test.env)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("InEnv(%q) = %+v, want %+v", test.env, got, test.want)
			}
		})
	}
}
