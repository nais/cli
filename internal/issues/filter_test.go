package issues

import (
	"reflect"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestParse(t *testing.T) {
	type want struct {
		filters gql.IssueFilter
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
				filters: gql.IssueFilter{
					Environments: []string{"x"},
				},
			},
		},
		{
			name:  "multiple filters",
			input: "environment=x,severity=CRITICAL",
			want: want{
				filters: gql.IssueFilter{
					Environments: []string{"x"},
					Severity:     "CRITICAL",
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
		{
			name:  "invalid filter value",
			input: "severity=marning",
			want: want{
				err: "invalid filter value: marning, valid values are: [CRITICAL WARNING TODO]",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseFilter(test.input)

			if test.want.err != "" {
				if err == nil {
					t.Errorf("parseFilter(%q) = %+v, want error %+v", test.input, err, test.want.err)
				}

				if err.Error() != test.want.err {
					t.Errorf("parseFilter(%q) = %+v, want error %+v", test.input, err, test.want.err)
				}
				return
			}

			if !reflect.DeepEqual(got, test.want.filters) {
				t.Errorf("parseFilter(%q): %+v, want = %+v", test.input, got, test.want.filters)
			}
		})
	}
}
