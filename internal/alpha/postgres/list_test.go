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
