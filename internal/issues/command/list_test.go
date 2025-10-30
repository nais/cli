package command

import (
	"testing"

	"github.com/nais/cli/internal/issues/command/flag"
)

func TestParseFilters(t *testing.T) {
	type want struct {
		filters *flag.Filters
		err     string
	}

	tests := []struct {
		name  string
		input string
		want  want
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
					Environment: "x",
					Severity:    "CRITICAL",
				},
			},
		},
		{
			name:  "unknown filter",
			input: "unknown=something",
			want: want{
				err: "unknown filter key: unknown",
			},
		},
		{
			name:  "malformed filter",
			input: "environment=x,severity",
			want: want{
				err: "incorrect filter: severity",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseFilter(test.input)

			if test.want.err != "" {
				if err == nil {
					t.Errorf("parseFilter(%q) = %+v, want error %+v", test.input, err, test.want.err)
				}

				if err.Error() != test.want.err {
					t.Errorf("parseFilter(%q) = %+v, want %+v", test.input, err, test.want.err)
				}
				return
			}

			if *got != *test.want.filters {
				t.Errorf("parseFilter(%q) = %+v, want %+v", test.input, got, test.want.filters)
			}
		})
	}
}
