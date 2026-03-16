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
	}

	for _, tt := range tests {
		tt := tt
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CountFlagOccurrences(tt.args, tt.short, tt.long)
			if got != tt.wantCount {
				t.Fatalf("CountFlagOccurrences() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}