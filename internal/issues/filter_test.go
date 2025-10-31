package issues

import (
	"reflect"
	"testing"

	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

func TestParse(t *testing.T) {
	type want struct {
		filters gql.IssueFilter
		err     string
	}

	tests := []struct {
		name  string
		input *flag.List
		want  want
	}{
		{
			name:  "single filter",
			input: &flag.List{Environment: "x"},
			want: want{
				filters: gql.IssueFilter{
					Environments: []string{"x"},
				},
			},
		},
		{
			name:  "multiple filters",
			input: &flag.List{Environment: "x", Severity: "CRITICAL"},
			want: want{
				filters: gql.IssueFilter{
					Environments: []string{"x"},
					Severity:     "CRITICAL",
				},
			},
		},
		{
			name:  "invalid filter value",
			input: &flag.List{Severity: "marning"},
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
					t.Errorf("parseFilter(%+v) = %+v, want error %+v", test.input, err, test.want.err)
				}

				if err.Error() != test.want.err {
					t.Errorf("parseFilter(%+v) = %+v, want error %+v", test.input, err, test.want.err)
				}
				return
			}

			if !reflect.DeepEqual(got, test.want.filters) {
				t.Errorf("parseFilter(%+v): %+v, want = %+v", test.input, got, test.want.filters)
			}
		})
	}
}
