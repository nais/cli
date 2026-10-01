package cloudsql

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix/output"
)

func TestInstancesFromTeam(t *testing.T) {
	const teamData = `{"sqlInstances":{"nodes":[
		{"name":"other","teamEnvironment":{"environment":{"name":"prod"}},"version":null,"highAvailability":true,"auditLog":null,"state":"STOPPED"},
		{"name":"legacy","teamEnvironment":{"environment":{"name":"dev"}},"version":"POSTGRES_14","highAvailability":false,"auditLog":{"logUrl":"https://example.test"},"state":"RUNNABLE"}
	]}}`
	var data gql.GetTeamCloudSQLInstancesTeam
	if err := json.Unmarshal([]byte(teamData), &data); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name         string
		environments []string
		want         []Instance
	}{
		{name: "sorted", want: []Instance{
			{Name: output.Link{Name: "legacy", URL: consoleBaseURL + "/team/my-team/dev/cloudsql/legacy"}, Type: "Cloud SQL", Environment: "dev", Version: "POSTGRES_14", Audit: boolPtr(true), State: State(gql.SqlInstanceStateRunnable)},
			{Name: output.Link{Name: "other", URL: consoleBaseURL + "/team/my-team/prod/cloudsql/other"}, Type: "Cloud SQL", Environment: "prod", HighAvailability: true, Audit: boolPtr(false), State: State(gql.SqlInstanceStateStopped)},
		}},
		{name: "environment filter", environments: []string{"dev"}, want: []Instance{
			{Name: output.Link{Name: "legacy", URL: consoleBaseURL + "/team/my-team/dev/cloudsql/legacy"}, Type: "Cloud SQL", Environment: "dev", Version: "POSTGRES_14", Audit: boolPtr(true), State: State(gql.SqlInstanceStateRunnable)},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := instancesFromTeam(data, "my-team", tt.environments); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("instancesFromTeam() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
func boolPtr(value bool) *bool { return &value }
