package postgres

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix/output"
)

func TestInstancesFromTeam(t *testing.T) {
	const teamData = `{
		"postgresBranches": {"nodes": [
			{"name":"main","teamEnvironment":{"environment":{"name":"dev"}},"postgres":{"name":"orders","majorVersion":"16","highAvailability":true},"state":"AVAILABLE"},
			{"name":"preview","teamEnvironment":{"environment":{"name":"prod"}},"postgres":{"name":"orders","majorVersion":"16","highAvailability":true},"state":"PROGRESSING"}
		]},
		"sqlInstances": {"nodes": [
			{"name":"legacy","teamEnvironment":{"environment":{"name":"dev"}},"version":"POSTGRES_14","highAvailability":false,"auditLog":{"logUrl":"https://example.test"},"state":"RUNNABLE"},
			{"name":"other","teamEnvironment":{"environment":{"name":"prod"}},"version":null,"highAvailability":true,"auditLog":null,"state":"STOPPED"}
		]}
	}`
	var data gql.GetTeamPostgresBranchesTeam
	if err := json.Unmarshal([]byte(teamData), &data); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		environments []string
		want         []Instance
	}{
		{
			name: "merged and sorted",
			want: []Instance{
				{Name: output.Link{Name: "legacy", URL: consoleBaseURL + "/team/my-team/dev/cloudsql/legacy"}, Type: "Cloud SQL", Environment: "dev", Version: "POSTGRES_14", Audit: new(true), State: State(gql.SqlInstanceStateRunnable)},
				{Name: output.Link{Name: "orders/main", URL: consoleBaseURL + "/team/my-team/dev/postgres/orders"}, Type: "PostgreSQL", Environment: "dev", Version: "16", HighAvailability: true, State: State(gql.PostgresBranchStateAvailable)},
				{Name: output.Link{Name: "orders/preview", URL: consoleBaseURL + "/team/my-team/prod/postgres/orders"}, Type: "PostgreSQL", Environment: "prod", Version: "16", HighAvailability: true, State: State(gql.PostgresBranchStateProgressing)},
				{Name: output.Link{Name: "other", URL: consoleBaseURL + "/team/my-team/prod/cloudsql/other"}, Type: "Cloud SQL", Environment: "prod", HighAvailability: true, Audit: new(false), State: State(gql.SqlInstanceStateStopped)},
			},
		},
		{
			name:         "environment filter applies to both providers",
			environments: []string{"dev"},
			want: []Instance{
				{Name: output.Link{Name: "legacy", URL: consoleBaseURL + "/team/my-team/dev/cloudsql/legacy"}, Type: "Cloud SQL", Environment: "dev", Version: "POSTGRES_14", Audit: new(true), State: State(gql.SqlInstanceStateRunnable)},
				{Name: output.Link{Name: "orders/main", URL: consoleBaseURL + "/team/my-team/dev/postgres/orders"}, Type: "PostgreSQL", Environment: "dev", Version: "16", HighAvailability: true, State: State(gql.PostgresBranchStateAvailable)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instancesFromTeam(data, "my-team", tt.environments)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("instancesFromTeam() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

//go:fix inline
func boolPtr(value bool) *bool { return new(value) }
