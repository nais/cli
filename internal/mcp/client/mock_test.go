package client

import (
	"context"
	"strings"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestMockClient_GetCurrentUser(t *testing.T) {
	client := NewMockClient(ScenarioHealthy)

	user, err := client.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Email != "mock-user@example.com" {
		t.Errorf("expected email mock-user@example.com, got %s", user.Email)
	}

	if user.IsAdmin {
		t.Error("expected non-admin user")
	}
}

func TestMockClient_GetUserTeams(t *testing.T) {
	client := NewMockClient(ScenarioHealthy)

	teams, err := client.GetUserTeams(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(teams))
	}

	// Check first team
	if teams[0].Team.Slug != "team-alpha" {
		t.Errorf("expected team-alpha, got %s", teams[0].Team.Slug)
	}

	if teams[0].Role != gql.TeamMemberRoleMember {
		t.Errorf("expected member role, got %s", teams[0].Role)
	}

	// Check second team
	if teams[1].Team.Slug != "team-beta" {
		t.Errorf("expected team-beta, got %s", teams[1].Team.Slug)
	}

	if teams[1].Role != gql.TeamMemberRoleOwner {
		t.Errorf("expected owner role, got %s", teams[1].Role)
	}
}

func TestMockClient_GetSchema(t *testing.T) {
	client := NewMockClient(ScenarioHealthy)

	schema, err := client.GetSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema == "" {
		t.Error("expected non-empty schema")
	}

	// Check schema contains expected types
	if !strings.Contains(schema, "type Query") {
		t.Error("expected schema to contain 'type Query'")
	}
	if !strings.Contains(schema, "type Team") {
		t.Error("expected schema to contain 'type Team'")
	}
	if !strings.Contains(schema, "type Application") {
		t.Error("expected schema to contain 'type Application'")
	}
}

func TestMockClient_GetConsoleURL(t *testing.T) {
	client := NewMockClient(ScenarioHealthy)

	url, err := client.GetConsoleURL(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://console.nav.cloud.nais.io"
	if url != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, url)
	}
}

func TestMockClient_SetScenario(t *testing.T) {
	client := NewMockClient(ScenarioHealthy)

	// Verify initial scenario
	if client.scenario != ScenarioHealthy {
		t.Errorf("expected ScenarioHealthy, got %s", client.scenario)
	}

	// Switch scenario
	client.SetScenario(ScenarioCrashLoop)

	if client.scenario != ScenarioCrashLoop {
		t.Errorf("expected ScenarioCrashLoop, got %s", client.scenario)
	}
}
