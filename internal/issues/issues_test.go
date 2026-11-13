package issues

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestPostgresIssueIdentifiesDatabase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct{ OperationName, Query string }
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.OperationName != "GetAllIssues" {
			t.Errorf("unexpected operation: %s", request.OperationName)
		}
		if strings.Contains(request.Query, "PostgresBranch") {
			t.Error("issue query must not depend on the retired branch model")
		}
		if !strings.Contains(request.Query, "... on PostgresIssue") || !strings.Contains(request.Query, "postgres {") {
			t.Error("issue query must select the affected Postgres database")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"team":{"issues":{"nodes":[{"__typename":"PostgresIssue","id":"issue-orders","severity":"CRITICAL","message":"Database failure","teamEnvironment":{"environment":{"name":"dev"}},"postgres":{"__typename":"Postgres","name":"orders"}}]}}}}`))
	}))
	defer server.Close()
	t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
	t.Setenv("NAIS_API_LOCAL_EMAIL", "probe@example.com")
	got, err := GetAll(t.Context(), "team", gql.IssueFilter{})
	if err != nil {
		t.Fatal(err)
	}
	want := []Issue{{ID: "issue-orders", Severity: Severity(gql.SeverityCritical), Environment: "dev", ResourceName: "orders", ResourceType: "Postgres", Message: "Database failure"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("issues = %+v, want %+v", got, want)
	}
}
