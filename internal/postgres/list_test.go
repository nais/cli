package postgres

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
	for _, host := range []string{"console.nav.cloud.nais.io", "console.dev-nais.cloud.nais.io"} {
		t.Run(host, func(t *testing.T) {
			want := []PostgresInstance{
				{Name: output.Link{Name: "orders", URL: "https://" + host + "/team/my-team/dev/postgres/orders"}, Environment: "dev", Version: "17", HighAvailability: true, ActiveBranch: "main"},
				{Name: output.Link{Name: "reports", URL: "https://" + host + "/team/my-team/prod/postgres/reports"}, Environment: "prod", Version: "16"},
			}
			if got := postgresesFromTeam(data, "my-team", host); !reflect.DeepEqual(got, want) {
				t.Errorf("postgresesFromTeam() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestGetTeamPostgresesUsesAuthenticatedConsoleHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}},"majorVersion":"18","highAvailability":true}],"pageInfo":{"hasNextPage":false}}}}}`))
	}))
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")
	t.Setenv("NAIS_API_LOCAL_HOST", host)

	got, err := GetTeamPostgreses(t.Context(), "my-team", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []PostgresInstance{
		{Name: output.Link{Name: "orders", URL: "https://" + host + "/team/my-team/dev/postgres/orders"}, Environment: "dev", Version: "18", HighAvailability: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetTeamPostgreses() = %#v, want %#v", got, want)
	}
}
