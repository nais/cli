package cliflags

import (
	"reflect"
	"testing"
)

func TestUniqueFlagValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		short    string
		long     string
		wantVals []string
	}{
		{
			name:     "no flags",
			args:     []string{"cmd", "sub"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{},
		},
		{
			name:     "short flag with separate value",
			args:     []string{"cmd", "-e", "dev-gcp"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{"dev-gcp"},
		},
		{
			name:     "long flag with separate value",
			args:     []string{"cmd", "--environment", "prod-gcp"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{"prod-gcp"},
		},
		{
			name:     "short equals form",
			args:     []string{"cmd", "-e=dev-gcp"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{"dev-gcp"},
		},
		{
			name:     "long equals form",
			args:     []string{"cmd", "--environment=prod-gcp"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{"prod-gcp"},
		},
		{
			name:     "mixed forms preserve first-seen order and dedupe",
			args:     []string{"cmd", "-e", "dev-gcp", "--environment=prod-gcp", "-e=dev-gcp", "--environment", "prod-gcp", "-e", "dev-fss"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{"dev-gcp", "prod-gcp", "dev-fss"},
		},
		{
			name:     "missing value is ignored",
			args:     []string{"cmd", "-e"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{},
		},
		{
			name:     "flag-like next arg is ignored",
			args:     []string{"cmd", "--environment", "--team", "nais"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{},
		},
		{
			name:     "empty equals value is ignored",
			args:     []string{"cmd", "--environment=", "-e="},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{},
		},
		{
			name:     "end-of-flags marker is not treated as value",
			args:     []string{"cmd", "-e", "--", "secret"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{},
		},
		{
			name:     "flags after end-of-flags marker are ignored",
			args:     []string{"cmd", "--", "-e", "dev-gcp", "--environment=prod-gcp"},
			short:    "-e",
			long:     "--environment",
			wantVals: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := UniqueFlagValues(tt.args, tt.short, tt.long)
			if !reflect.DeepEqual(got, tt.wantVals) {
				t.Fatalf("UniqueFlagValues() = %#v, want %#v", got, tt.wantVals)
			}
		})
	}
}

func TestCountFlagOccurrences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		short     string
		long      string
		wantCount int
	}{
		{
			name:      "no flags",
			args:      []string{"cmd", "sub"},
			short:     "-e",
			long:      "--environment",
			wantCount: 0,
		},
		{
			name:      "all supported forms counted",
			args:      []string{"cmd", "-e", "dev-gcp", "--environment", "prod-gcp", "-e=dev-fss", "--environment=prod-fss"},
			short:     "-e",
			long:      "--environment",
			wantCount: 4,
		},
		{
			name:      "duplicates still count as separate occurrences",
			args:      []string{"cmd", "-e", "dev-gcp", "--environment=dev-gcp", "-e", "dev-gcp"},
			short:     "-e",
			long:      "--environment",
			wantCount: 3,
		},
		{
			name:      "missing and invalid values ignored",
			args:      []string{"cmd", "-e", "--environment", "--environment=", "-e=", "--environment", "--team", "nais"},
			short:     "-e",
			long:      "--environment",
			wantCount: 0,
		},
		{
			name:      "flags after end-of-flags marker are ignored",
			args:      []string{"cmd", "--", "-e", "dev-gcp", "--environment=prod-gcp"},
			short:     "-e",
			long:      "--environment",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CountFlagOccurrences(tt.args, tt.short, tt.long)
			if got != tt.wantCount {
				t.Fatalf("CountFlagOccurrences() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestFirstFlagValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  []string
		short string
		long  string
		want  string
	}{
		{
			name:  "no matching flag",
			args:  []string{"cmd", "sub"},
			short: "-t",
			long:  "--team",
			want:  "",
		},
		{
			name:  "short flag with separate value",
			args:  []string{"cmd", "-t", "nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "long flag with separate value",
			args:  []string{"cmd", "--team", "nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "short equals form",
			args:  []string{"cmd", "-t=nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "long equals form",
			args:  []string{"cmd", "--team=nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "first occurrence wins",
			args:  []string{"cmd", "--team=nais", "-t", "other"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "missing value with no later value returns empty",
			args:  []string{"cmd", "--team"},
			short: "-t",
			long:  "--team",
			want:  "",
		},
		{
			name:  "flag-like next arg with no later value returns empty",
			args:  []string{"cmd", "--team", "--environment", "dev"},
			short: "-t",
			long:  "--team",
			want:  "",
		},
		{
			name:  "missing value continues scanning to later valid value",
			args:  []string{"cmd", "--team", "-x", "--team=nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "flag-like next arg continues scanning to later valid value",
			args:  []string{"cmd", "--team", "--environment", "dev", "-t", "nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "empty equals value ignored and later value used",
			args:  []string{"cmd", "--team=", "-t", "nais"},
			short: "-t",
			long:  "--team",
			want:  "nais",
		},
		{
			name:  "flags after end-of-flags marker ignored",
			args:  []string{"cmd", "--", "--team", "nais"},
			short: "-t",
			long:  "--team",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FirstFlagValue(tt.args, tt.short, tt.long)
			if got != tt.want {
				t.Fatalf("FirstFlagValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasSubCommandPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		parent  string
		subs    []string
		wantHas bool
	}{
		{
			name:    "matches parent and subcommand",
			args:    []string{"nais", "app", "restart"},
			parent:  "app",
			subs:    []string{"restart"},
			wantHas: true,
		},
		{
			name:    "matches with flag between parent and subcommand",
			args:    []string{"nais", "job", "--team", "my-team", "trigger"},
			parent:  "job",
			subs:    []string{"trigger"},
			wantHas: true,
		},
		{
			name:    "matches one of many subcommands",
			args:    []string{"nais", "valkey", "list"},
			parent:  "valkey",
			subs:    []string{"credentials", "delete", "get", "list", "update"},
			wantHas: true,
		},
		{
			name:    "returns false for different subcommand",
			args:    []string{"nais", "app", "list"},
			parent:  "app",
			subs:    []string{"restart"},
			wantHas: false,
		},
		{
			name:    "returns false when token appears as positional argument",
			args:    []string{"nais", "opensearch", "create", "delete"},
			parent:  "opensearch",
			subs:    []string{"credentials", "delete", "get", "list", "update"},
			wantHas: false,
		},
		{
			name:    "stops at end of flags marker",
			args:    []string{"nais", "app", "--", "restart"},
			parent:  "app",
			subs:    []string{"restart"},
			wantHas: false,
		},
		{
			name:    "returns false with no subcommands provided",
			args:    []string{"nais", "app", "restart"},
			parent:  "app",
			subs:    nil,
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := HasSubCommandPath(tt.args, tt.parent, tt.subs...)
			if got != tt.wantHas {
				t.Fatalf("HasSubCommandPath() = %v, want %v", got, tt.wantHas)
			}
		})
	}
}
