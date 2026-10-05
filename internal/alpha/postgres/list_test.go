package postgres

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix/output"
)

func TestPostgresesFromTeam(t *testing.T) {
	const teamData = `{"postgreses":{"nodes":[
		{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}},"majorVersion":"17","highAvailability":true,"activeBranch":{"name":"main"}},
		{"name":"reports","teamEnvironment":{"environment":{"name":"prod"}},"majorVersion":"16","highAvailability":false}
	]}}`
	var data gql.GetTeamPostgresesTeam
	if err := json.Unmarshal([]byte(teamData), &data); err != nil {
		t.Fatal(err)
	}
	want := []PostgresInstance{
		{Name: output.Link{Name: "orders", URL: consoleBaseURL + "/team/my-team/dev/postgres/orders"}, Environment: "dev", Version: "17", HighAvailability: true, ActiveBranch: "main"},
		{Name: output.Link{Name: "reports", URL: consoleBaseURL + "/team/my-team/prod/postgres/reports"}, Environment: "prod", Version: "16"},
	}
	if got := postgresesFromTeam(data, "my-team"); !reflect.DeepEqual(got, want) {
		t.Errorf("postgresesFromTeam() = %#v, want %#v", got, want)
	}
}

func TestInstancesFromTeam(t *testing.T) {
	const teamData = `{"postgresBranches":{"nodes":[
		{"name":"preview","teamEnvironment":{"environment":{"name":"prod"}},"postgres":{"name":"orders","majorVersion":"16","highAvailability":true},"state":"PROGRESSING"},
		{"name":"main","teamEnvironment":{"environment":{"name":"dev"}},"postgres":{"name":"orders","majorVersion":"16","highAvailability":true},"state":"AVAILABLE"}
	]}}`
	var data gql.GetTeamPostgresBranchesTeam
	if err := json.Unmarshal([]byte(teamData), &data); err != nil {
		t.Fatal(err)
	}
	dev := Instance{Name: output.Link{Name: "orders/main", URL: consoleBaseURL + "/team/my-team/dev/postgres/orders"}, Type: "PostgreSQL", Environment: "dev", Version: "16", HighAvailability: true, State: State(gql.PostgresBranchStateAvailable)}
	prod := Instance{Name: output.Link{Name: "orders/preview", URL: consoleBaseURL + "/team/my-team/prod/postgres/orders"}, Type: "PostgreSQL", Environment: "prod", Version: "16", HighAvailability: true, State: State(gql.PostgresBranchStateProgressing)}
	for _, tt := range []struct {
		name         string
		environments []string
		want         []Instance
	}{
		{name: "sorted branches", want: []Instance{dev, prod}},
		{name: "environment filter", environments: []string{"dev"}, want: []Instance{dev}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := instancesFromTeam(data, "my-team", tt.environments); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("instancesFromTeam() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
