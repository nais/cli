package tools

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateQueryDocument(t *testing.T) {
	schema := loadRealSchema(t)

	tests := []struct {
		name      string
		query     string
		wantValid bool
		wantErr   string
	}{
		{
			name:      "single query is allowed",
			query:     `query { team(slug: "a") { slug } }`,
			wantValid: true,
		},
		{
			name:      "named fragment without secrets is allowed",
			query:     `query { team(slug: "a") { ...TeamFields } } fragment TeamFields on Team { slug purpose }`,
			wantValid: true,
		},
		{
			name:    "mutation is refused",
			query:   `mutation { restartApplication(input: {name: "a", teamSlug: "b", environmentName: "c"}) { __typename } }`,
			wantErr: "only query operations are allowed",
		},
		{
			name: "mutation after query in same document is refused",
			query: `
				query Harmless { team(slug: "a") { slug } }
				mutation Evil { restartApplication(input: {name: "a", teamSlug: "b", environmentName: "c"}) { __typename } }
			`,
			wantErr: "only a single operation is allowed",
		},
		{
			name: "multiple queries are refused",
			query: `
				query A { team(slug: "a") { slug } }
				query B { team(slug: "b") { slug } }
			`,
			wantErr: "only a single operation is allowed",
		},
		{
			name:    "secrets accessed through named fragment are refused",
			query:   `query { team(slug: "a") { ...Innocent } } fragment Innocent on Team { secrets(first: 1) { nodes { name } } }`,
			wantErr: "field 'secrets' returns type 'SecretConnection'",
		},
		{
			name: "secrets accessed through nested named fragments are refused",
			query: `
				query { team(slug: "a") { ...Outer } }
				fragment Outer on Team { slug ...Inner }
				fragment Inner on Team { secrets(first: 1) { nodes { name } } }
			`,
			wantErr: "field 'secrets' returns type 'SecretConnection'",
		},
		{
			// https://github.com/nais/cli/issues/803
			name:    "deployment key accessed through named fragment is refused",
			query:   `query { team(slug: "a") { ...TeamFields } } fragment TeamFields on Team { deploymentKey { key } }`,
			wantErr: "field 'deploymentKey' returns type 'DeploymentKey'",
		},
		{
			name:    "fragment on forbidden type is refused",
			query:   `query { team(slug: "a") { secrets(first: 1) { nodes { ...S } } } } fragment S on Secret { name }`,
			wantErr: "MCP security policy",
		},
		{
			name:    "depth hidden in named fragments is counted",
			query:   `query { team(slug: "a") { ...F1 } }` + deepFragments(20),
			wantErr: "exceeds maximum allowed depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := validateQueryDocument(schema, tt.query)
			if res.valid != tt.wantValid {
				t.Fatalf("valid = %v, want %v (error: %q)", res.valid, tt.wantValid, res.error)
			}
			if tt.wantErr != "" && !strings.Contains(res.error, tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", res.error, tt.wantErr)
			}
		})
	}
}

// deepFragments builds a chain of n fragments on Team, each adding nesting via
// team.environments.team, so the total depth is only visible when fragments are resolved.
func deepFragments(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		if i < n {
			fmt.Fprintf(&b, "\nfragment F%d on Team { slug environments { team { ...F%d } } }", i, i+1)
		} else {
			fmt.Fprintf(&b, "\nfragment F%d on Team { slug }", i)
		}
	}
	return b.String()
}
