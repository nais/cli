package postgres

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix/output"
)

func TestInstancesFromTeam(t *testing.T) {
	const teamData = `{"postgresBranches":{"nodes":[
		{"name":"preview","teamEnvironment":{"environment":{"name":"prod"}},"postgres":{"name":"orders","majorVersion":"16","highAvailability":true},"state":"PROGRESSING"},
		{"name":"main","teamEnvironment":{"environment":{"name":"dev"}},"postgres":{"name":"orders","majorVersion":"16","highAvailability":true},"state":"AVAILABLE"}
	]}}`
	var data gql.GetTeamPostgresBranchesAlphaTeam
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
