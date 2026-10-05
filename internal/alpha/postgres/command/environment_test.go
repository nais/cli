package command

import (
	"slices"
	"testing"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/naistrix/output"
)

func TestPostgresChoices(t *testing.T) {
	instances := []postgres.Instance{
		{Name: output.Link{Name: "relay-test/main"}, Environment: "dev-gcp"},
		{Name: output.Link{Name: "relay-test/restore"}, Environment: "dev-gcp"},
		{Name: output.Link{Name: "relay-test/main"}, Environment: "prod-gcp"},
		{Name: output.Link{Name: "other/main"}, Environment: "prod-gcp"},
	}
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
	if got, want := postgresBranches(instances, "relay-test", "dev-gcp"), []string{"main", "restore"}; !slices.Equal(got, want) {
		t.Errorf("dev branches = %v, want %v", got, want)
	}
	if got, want := postgresBranches(instances, "relay-test", "prod-gcp"), []string{"main"}; !slices.Equal(got, want) {
		t.Errorf("prod branches = %v, want %v", got, want)
	}
	if got := postgresBranches(instances, "other", "dev-gcp"); len(got) != 0 {
		t.Errorf("other Postgres branches in dev = %v, want none", got)
	}
	if got, help := branchSuggestions(instances, "relay-test", ""); len(got) != 0 || help == "" {
		t.Errorf("ambiguous environment completed branches: %v, %q", got, help)
	}
	if got, _ := branchSuggestions(instances, "other", ""); !slices.Equal(got, []string{"main"}) {
		t.Errorf("single-environment branches = %v, want [main]", got)
	}
	if got, _ := branchSuggestions(instances, "relay-test", "prod-gcp"); !slices.Equal(got, []string{"main"}) {
		t.Errorf("explicit-environment branches = %v, want [main]", got)
	}
}
