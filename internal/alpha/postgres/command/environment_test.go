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
	for _, tc := range []struct {
		name, environment string
		want              []string
	}{
		{name: "all instances", want: []string{"other", "relay-test"}},
		{name: "instances in dev", environment: "dev-gcp", want: []string{"relay-test"}},
		{name: "instances in prod", environment: "prod-gcp", want: []string{"other", "relay-test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := postgresNames(instances, tc.environment); !slices.Equal(got, tc.want) {
				t.Errorf("names = %v, want %v", got, tc.want)
			}
		})
	}
	if got, want := postgresEnvironments(instances, "relay-test"), []string{"dev-gcp", "prod-gcp"}; !slices.Equal(got, want) {
		t.Errorf("environments = %v, want %v", got, want)
	}
	if got := postgresEnvironments(instances, "missing"); len(got) != 0 {
		t.Errorf("environments for missing Postgres = %v, want none", got)
	}
}
