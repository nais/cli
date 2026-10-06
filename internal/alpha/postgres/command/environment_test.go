package command

import (
	"slices"
	"testing"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/naistrix/output"
)

func TestPostgresChoices(t *testing.T) {
	postgreses := []postgres.PostgresInstance{
		{Name: output.Link{Name: "relay-test"}, Environment: "dev-gcp"},
		{Name: output.Link{Name: "relay-test"}, Environment: "prod-gcp"},
		{Name: output.Link{Name: "other"}, Environment: "prod-gcp"},
		{Name: output.Link{Name: "branchless"}, Environment: "dev-gcp"},
	}
	for _, tc := range []struct {
		name, environment string
		want              []string
	}{
		{name: "all instances", want: []string{"branchless", "other", "relay-test"}},
		{name: "instances in dev", environment: "dev-gcp", want: []string{"branchless", "relay-test"}},
		{name: "instances in prod", environment: "prod-gcp", want: []string{"other", "relay-test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := postgresNames(postgreses, tc.environment); !slices.Equal(got, tc.want) {
				t.Errorf("names = %v, want %v", got, tc.want)
			}
		})
	}
	if got, want := postgresEnvironments(postgreses, "relay-test"), []string{"dev-gcp", "prod-gcp"}; !slices.Equal(got, want) {
		t.Errorf("environments = %v, want %v", got, want)
	}
	if got := postgresEnvironments(postgreses, "missing"); len(got) != 0 {
		t.Errorf("environments for missing Postgres = %v, want none", got)
	}
	if got, want := postgresEnvironments(postgreses, "branchless"), []string{"dev-gcp"}; !slices.Equal(got, want) {
		t.Errorf("branchless environments = %v, want %v", got, want)
	}
	if env, help := branchCompletionEnvironment(postgreses, "relay-test", ""); env != "" || help == "" {
		t.Errorf("ambiguous Postgres environment: %q, %q", env, help)
	}
	if env, help := branchCompletionEnvironment(postgreses, "other", ""); env != "prod-gcp" || help != "" {
		t.Errorf("single Postgres environment: %q, %q", env, help)
	}
	if env, help := branchCompletionEnvironment(postgreses, "relay-test", "dev-gcp"); env != "dev-gcp" || help != "" {
		t.Errorf("explicit Postgres environment: %q, %q", env, help)
	}
	if got, help := branchSuggestions([]postgres.Branch{{Name: "restore"}, {Name: "main"}}); !slices.Equal(got, []string{"main", "restore"}) || help == "" {
		t.Errorf("branch suggestions = %v, %q", got, help)
	}
	if got, help := branchSuggestions([]postgres.Branch{{Name: "main"}}); !slices.Equal(got, []string{"main"}) || help != "" {
		t.Errorf("single branch suggestion = %v, %q", got, help)
	}
}
