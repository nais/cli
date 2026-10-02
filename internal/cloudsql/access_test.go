package cloudsql

import (
	"strings"
	"testing"
)

func TestAccessStatement(t *testing.T) {
	tests := []struct {
		name, template, schema, group, want string
	}{
		{
			name:     "individual IAM users by default",
			template: grantSelectPrivs,
			schema:   "public",
			want:     `GRANT USAGE ON SCHEMA "public" TO cloudsqliamuser;`,
		},
		{
			name:     "IAM group receives access to existing tables",
			template: grantSelectPrivs,
			schema:   "public",
			group:    "dapla-metadata-developers@groups.ssb.no",
			want:     `GRANT SELECT ON ALL TABLES IN SCHEMA "public" TO "dapla-metadata-developers@groups.ssb.no";`,
		},
		{
			name:     "revoke from the requested group",
			template: revokeAllPrivs,
			schema:   "my schema",
			group:    "team@groups.ssb.no",
			want:     `REVOKE ALL ON ALL TABLES IN SCHEMA "my schema" FROM "team@groups.ssb.no";`,
		},
		{
			name:     "quote identifiers rather than interpreting them as SQL",
			template: grantSelectPrivs,
			schema:   `a"b`,
			group:    `x"y@groups.ssb.no`,
			want:     `GRANT USAGE ON SCHEMA "a""b" TO "x""y@groups.ssb.no";`,
		},
		{
			name:     "do not expand placeholders inside quoted identifiers",
			template: grantSelectPrivs,
			schema:   "$role",
			group:    "$schema@groups.ssb.no",
			want:     `GRANT USAGE ON SCHEMA "$role" TO "$schema@groups.ssb.no";`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessStatement(tt.template, tt.schema, tt.group)
			if !strings.Contains(got, tt.want) {
				t.Errorf("statement %q does not contain %q", got, tt.want)
			}
		})
	}
}
