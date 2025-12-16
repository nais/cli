package tools

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/nais/cli/internal/mcp/client"
	gql "github.com/nais/cli/internal/naisapi/gql"
)

// mockClientWithCounter is a mock client that counts GetSchema calls.
type mockClientWithCounter struct {
	callCount int
	mu        sync.Mutex
	schema    string
}

func (m *mockClientWithCounter) GetCurrentUser(ctx context.Context) (*client.User, error) {
	return nil, nil
}

func (m *mockClientWithCounter) GetUserTeams(ctx context.Context) ([]gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMember, error) {
	return nil, nil
}

func (m *mockClientWithCounter) GetSchema(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.schema, nil
}

func (m *mockClientWithCounter) GetConsoleURL(ctx context.Context) (string, error) {
	return "https://console.example.com", nil
}

func (m *mockClientWithCounter) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestSchemaCaching(t *testing.T) {
	mockClient := &mockClientWithCounter{
		schema: `type Query { test: String }`,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	ctx := &toolContext{
		client: mockClient,
		logger: logger,
	}

	reqCtx := context.Background()

	// First call should fetch the schema
	schema1, err := ctx.getCachedSchema(reqCtx)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	if schema1 == "" {
		t.Error("expected non-empty schema")
	}

	if mockClient.GetCallCount() != 1 {
		t.Errorf("expected 1 call to GetSchema, got %d", mockClient.GetCallCount())
	}

	// Second call should use cache
	schema2, err := ctx.getCachedSchema(reqCtx)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if schema2 != schema1 {
		t.Error("cached schema should be identical")
	}

	if mockClient.GetCallCount() != 1 {
		t.Errorf("expected still 1 call to GetSchema (cached), got %d", mockClient.GetCallCount())
	}

	// Third call should still use cache
	schema3, err := ctx.getCachedSchema(reqCtx)
	if err != nil {
		t.Fatalf("third call failed: %v", err)
	}

	if schema3 != schema1 {
		t.Error("cached schema should be identical")
	}

	if mockClient.GetCallCount() != 1 {
		t.Errorf("expected still 1 call to GetSchema (cached), got %d", mockClient.GetCallCount())
	}
}

func TestSchemaCaching_Concurrent(t *testing.T) {
	mockClient := &mockClientWithCounter{
		schema: `type Query { test: String }`,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	ctx := &toolContext{
		client: mockClient,
		logger: logger,
	}

	reqCtx := context.Background()

	// Call getCachedSchema concurrently from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := ctx.getCachedSchema(reqCtx)
			if err != nil {
				t.Errorf("concurrent call failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// Even with concurrent calls, GetSchema should only be called once
	if mockClient.GetCallCount() != 1 {
		t.Errorf("expected 1 call to GetSchema despite concurrent access, got %d", mockClient.GetCallCount())
	}
}

func TestSchemaCaching_BuiltinScalarsRemoved(t *testing.T) {
	mockClient := &mockClientWithCounter{
		schema: `
scalar Boolean
scalar String
scalar Int

type Query {
	test: String
}`,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	ctx := &toolContext{
		client: mockClient,
		logger: logger,
	}

	reqCtx := context.Background()

	schema, err := ctx.getCachedSchema(reqCtx)
	if err != nil {
		t.Fatalf("getCachedSchema failed: %v", err)
	}

	// Built-in scalars should be removed from cached schema
	if contains(schema, "scalar Boolean") {
		t.Error("cached schema should not contain 'scalar Boolean'")
	}
	if contains(schema, "scalar String") {
		t.Error("cached schema should not contain 'scalar String'")
	}
	if contains(schema, "scalar Int") {
		t.Error("cached schema should not contain 'scalar Int'")
	}

	// But the Query type should still be there
	if !contains(schema, "type Query") {
		t.Error("cached schema should contain 'type Query'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
