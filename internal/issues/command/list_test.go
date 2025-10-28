package command

import (
	"fmt"
	"testing"

	"github.com/nais/cli/internal/issues/command/flag"
)

func TestParseFilters(t *testing.T) {
	type want struct {
		filters *flag.Filters
		err     error
	}

	tests := []struct {
		name  string
		input string
		want
	}{
		{
			name:  "single filter",
			input: "environment=x",
			want: want{
				filters: &flag.Filters{
					Environment: "x",
				},
			},
		},
		{
			name:  "multiple filters",
			input: "environment=x,severity=CRITICAL",
			want: want{
				filters: &flag.Filters{
					Environment: "y",
					Severity:    "CRITICAL",
				},
			},
		},
		{
			name:  "unknown filter",
			input: "unknown=something",
			want: want{
				err: fmt.Errorf("unknown filter key: unknown"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseFilter(test.input)
			if err != test.err {
				t.Errorf("parseFilter(%q) = %+v, want %+v", test.input, err, test.err)
			}
			if *got != *test.want.filters {
				t.Errorf("parseFilter(%q) = %+v, want %+v", test.input, got, test.want)
			}
		})
	}
}
