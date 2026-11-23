package activity

import (
	"encoding/json"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestPostgresActivityEntriesCanBeListed(t *testing.T) {
	const response = `{"nodes":[
		{"__typename":"PostgresCreatedActivityLogEntry","actor":"user@example.com","createdAt":"2026-10-05T08:26:46Z","message":"Created Postgres","environmentName":"dev-gcp","resourceType":"POSTGRES","resourceName":"relay-test"},
		{"__typename":"PostgresUpdatedActivityLogEntry","actor":"user@example.com","createdAt":"2026-10-05T08:27:42Z","message":"Updated Postgres","environmentName":"dev-gcp","resourceType":"POSTGRES","resourceName":"relay-test"},
		{"__typename":"PostgresDeletedActivityLogEntry","actor":"user@example.com","createdAt":"2026-10-05T08:31:16Z","message":"Deleted Postgres","environmentName":"dev-gcp","resourceType":"POSTGRES","resourceName":"relay-test"}
	]}`
	var entries gql.GetTeamActivityTeamActivityLogActivityLogEntryConnection
	if err := json.Unmarshal([]byte(response), &entries); err != nil {
		t.Fatalf("decoding Postgres activity: %v", err)
	}
	wantMessages := []string{"Created Postgres", "Updated Postgres", "Deleted Postgres"}
	if got := len(entries.Nodes); got != len(wantMessages) {
		t.Fatalf("got %d entries, want %d", got, len(wantMessages))
	}
	for i, want := range wantMessages {
		if got := entries.Nodes[i].GetMessage(); got != want {
			t.Errorf("entry %d: got %q, want %q", i, got, want)
		}
	}
}
