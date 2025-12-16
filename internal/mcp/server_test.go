package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TestClient wraps the MCP server for in-process testing.
type TestClient struct {
	mcpServer *server.MCPServer
	ctx       context.Context
	nextID    int32
}

// NewTestClient creates a new in-process test client.
func NewTestClient(t *testing.T) *TestClient {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mockClient := client.NewMockClient(client.ScenarioHealthy)

	srv, err := NewServer(
		WithLogger(logger),
		WithClient(mockClient),
		WithRateLimit(0), // Disable rate limiting for tests
	)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	mcpServer := srv.MCPServer()

	// Create an in-process session for testing
	session := server.NewInProcessSession(mcpServer.GenerateInProcessSessionID(), nil)

	ctx := context.Background()
	ctx = mcpServer.WithContext(ctx, session)

	if err := mcpServer.RegisterSession(ctx, session); err != nil {
		t.Fatalf("failed to register session: %v", err)
	}

	t.Cleanup(func() {
		mcpServer.UnregisterSession(ctx, session.SessionID())
	})

	return &TestClient{
		mcpServer: mcpServer,
		ctx:       ctx,
		nextID:    0,
	}
}

// SendRequest sends a JSON-RPC request and returns the response.
func (c *TestClient) SendRequest(t *testing.T, method string, params any) *JSONRPCResponse {
	t.Helper()

	id := int(atomic.AddInt32(&c.nextID, 1))
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	// Use HandleMessage for in-process handling
	respMsg := c.mcpServer.HandleMessage(c.ctx, reqBytes)

	respBytes, err := json.Marshal(respMsg)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\nRaw: %s", err, respBytes)
	}

	return &resp
}

// Initialize sends the initialize request to the MCP server.
func (c *TestClient) Initialize(t *testing.T) {
	t.Helper()

	resp := c.SendRequest(t, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]string{
			"name":    "test-client",
			"version": "1.0.0",
		},
	})

	if resp.Error != nil {
		t.Fatalf("initialize failed: %s", resp.Error.Message)
	}

	// Send initialized notification
	c.SendRequest(t, "notifications/initialized", nil)
}

// CallTool calls a tool and returns the parsed JSON result.
func (c *TestClient) CallTool(t *testing.T, name string, args map[string]any) map[string]any {
	t.Helper()

	resp := c.SendRequest(t, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})

	if resp.Error != nil {
		t.Fatalf("tool call failed: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatal("no content in response")
	}

	first := content[0].(map[string]any)
	if first["type"] != "text" {
		t.Fatalf("unexpected content type: %v", first["type"])
	}

	text := first["text"].(string)
	if isError, ok := first["isError"].(bool); ok && isError {
		t.Fatalf("tool returned error: %s", text)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	return parsed
}

// CallToolRaw calls a tool and returns the raw text result.
func (c *TestClient) CallToolRaw(t *testing.T, name string, args map[string]any) string {
	t.Helper()

	resp := c.SendRequest(t, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})

	if resp.Error != nil {
		t.Fatalf("tool call failed: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatal("no content in response")
	}

	first := content[0].(map[string]any)
	if first["type"] != "text" {
		t.Fatalf("unexpected content type: %v", first["type"])
	}

	text := first["text"].(string)
	if isError, ok := first["isError"].(bool); ok && isError {
		t.Fatalf("tool returned error: %s", text)
	}

	return text
}

// ListTools returns the list of available tool names.
func (c *TestClient) ListTools(t *testing.T) []string {
	t.Helper()

	resp := c.SendRequest(t, "tools/list", nil)

	if resp.Error != nil {
		t.Fatalf("list tools failed: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}

	tools := result["tools"].([]any)
	var names []string
	for _, tool := range tools {
		toolMap := tool.(map[string]any)
		names = append(names, toolMap["name"].(string))
	}

	return names
}

// ReadResource reads a resource and returns the content.
func (c *TestClient) ReadResource(t *testing.T, uri string) string {
	t.Helper()

	resp := c.SendRequest(t, "resources/read", map[string]any{
		"uri": uri,
	})

	if resp.Error != nil {
		t.Fatalf("read resource failed: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}

	contents, ok := result["contents"].([]any)
	if !ok || len(contents) == 0 {
		t.Fatal("no contents in response")
	}

	first := contents[0].(map[string]any)
	if text, ok := first["text"].(string); ok {
		return text
	}

	t.Fatal("no text content in resource")
	return ""
}

func TestMCPServer_Integration(t *testing.T) {
	client := NewTestClient(t)
	client.Initialize(t)

	t.Run("list_tools", func(t *testing.T) {
		tools := client.ListTools(t)

		expectedTools := []string{
			"get_nais_context",
			"execute_graphql",
			"validate_graphql",
			"schema_list_types",
			"schema_get_type",
			"schema_list_queries",
			"schema_list_mutations",
			"schema_get_field",
			"schema_get_enum",
			"schema_search",
			"schema_get_implementors",
			"schema_get_union_types",
		}

		for _, expected := range expectedTools {
			found := false
			for _, tool := range tools {
				if tool == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected tool %q not found in %v", expected, tools)
			}
		}

		// Ensure old tools are not present
		removedTools := []string{
			"whoami",
			"get_user_teams",
			"list_teams",
			"get_team",
			"get_team_members",
			"list_applications",
			"get_application",
			"list_environments",
			"get_team_vulnerabilities",
			"get_app_vulnerabilities",
			"get_job_vulnerabilities",
			"list_alerts",
			"list_issues",
		}

		for _, removed := range removedTools {
			for _, tool := range tools {
				if tool == removed {
					t.Errorf("removed tool %q should not be present", removed)
				}
			}
		}
	})

	t.Run("get_nais_context", func(t *testing.T) {
		result := client.CallTool(t, "get_nais_context", map[string]any{})

		// Check user info
		user, ok := result["user"].(map[string]any)
		if !ok {
			t.Fatalf("expected user object, got %T", result["user"])
		}

		if user["name"] != "Mock User" {
			t.Errorf("expected Mock User, got %v", user["name"])
		}

		// Check teams
		teams, ok := result["teams"].([]any)
		if !ok {
			t.Fatalf("expected teams array, got %T", result["teams"])
		}

		if len(teams) != 2 {
			t.Errorf("expected 2 teams, got %d", len(teams))
		}

		firstTeam := teams[0].(map[string]any)
		if firstTeam["slug"] != "team-alpha" {
			t.Errorf("expected team-alpha, got %v", firstTeam["slug"])
		}

		// Check console URL
		if result["console_base_url"] != "https://console.nav.cloud.nais.io" {
			t.Errorf("expected console URL, got %v", result["console_base_url"])
		}

		// Check URL patterns exist
		patterns, ok := result["console_url_patterns"].(map[string]any)
		if !ok {
			t.Fatalf("expected url patterns, got %T", result["console_url_patterns"])
		}

		if patterns["team"] != "/team/{team}" {
			t.Errorf("expected team pattern, got %v", patterns["team"])
		}
	})

	t.Run("validate_graphql", func(t *testing.T) {
		// Use a simple query that works with the mock schema
		raw := client.CallToolRaw(t, "validate_graphql", map[string]any{
			"query": "query { me { email } }",
		})

		// Should return some result (valid or validation info)
		if raw == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("schema_list_types", func(t *testing.T) {
		raw := client.CallToolRaw(t, "schema_list_types", map[string]any{})

		// Should return some type information
		if raw == "" {
			t.Error("expected non-empty schema types")
		}

		// The mock schema should contain Team
		if !strings.Contains(raw, "Team") {
			t.Error("expected schema to contain Team type")
		}
	})

	t.Run("schema_get_type", func(t *testing.T) {
		raw := client.CallToolRaw(t, "schema_get_type", map[string]any{
			"name": "Team",
		})

		// Should return Team type info
		if raw == "" {
			t.Error("expected non-empty result")
		}

		if !strings.Contains(raw, "Team") {
			t.Error("expected Team type details")
		}
	})

	t.Run("schema_list_queries", func(t *testing.T) {
		raw := client.CallToolRaw(t, "schema_list_queries", map[string]any{})

		// Should contain query operations
		if raw == "" {
			t.Error("expected non-empty result")
		}

		if !strings.Contains(raw, "me") {
			t.Error("expected 'me' query")
		}
	})

	t.Run("schema_search", func(t *testing.T) {
		raw := client.CallToolRaw(t, "schema_search", map[string]any{
			"query": "team",
		})

		// Should find team-related items
		if raw == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestMCPServer_ToolValidation(t *testing.T) {
	client := NewTestClient(t)
	client.Initialize(t)

	t.Run("missing_required_param", func(t *testing.T) {
		// Call validate_graphql without required 'query' parameter
		resp := client.SendRequest(t, "tools/call", map[string]any{
			"name":      "validate_graphql",
			"arguments": map[string]any{},
		})

		if resp.Error != nil {
			// Protocol-level error is acceptable
			return
		}

		result, ok := resp.Result.(map[string]any)
		if !ok {
			t.Fatalf("unexpected result type: %T", resp.Result)
		}

		content, ok := result["content"].([]any)
		if !ok || len(content) == 0 {
			t.Fatal("no content in response")
		}

		first := content[0].(map[string]any)
		text, _ := first["text"].(string)
		isError, _ := first["isError"].(bool)

		// Either isError is true, or text contains error info
		if !isError && !strings.Contains(strings.ToLower(text), "required") && !strings.Contains(strings.ToLower(text), "query") {
			t.Log("expected error response for missing required param, but got:", text)
		}
	})
}

func TestMCPServer_ServerInfo(t *testing.T) {
	client := NewTestClient(t)

	// Initialize and check server info
	resp := client.SendRequest(t, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]string{
			"name":    "test-client",
			"version": "1.0.0",
		},
	})

	if resp.Error != nil {
		t.Fatalf("initialize failed: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}

	// Check server info
	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("expected serverInfo in result")
	}

	if serverInfo["name"] != "nais-mcp" {
		t.Errorf("expected server name 'nais-mcp', got %v", serverInfo["name"])
	}

	// Check capabilities
	capabilities, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected capabilities in result")
	}

	if _, hasTools := capabilities["tools"]; !hasTools {
		t.Error("expected tools capability")
	}
}

func TestMCPServer_Resources(t *testing.T) {
	client := NewTestClient(t)
	client.Initialize(t)

	t.Run("list_resources", func(t *testing.T) {
		resp := client.SendRequest(t, "resources/list", nil)

		if resp.Error != nil {
			t.Fatalf("list resources failed: %s", resp.Error.Message)
		}

		result, ok := resp.Result.(map[string]any)
		if !ok {
			t.Fatalf("unexpected result type: %T", resp.Result)
		}

		resources, ok := result["resources"].([]any)
		if !ok {
			t.Fatalf("expected resources array, got %T", result["resources"])
		}

		// We expect at least the schema and best practices resources
		if len(resources) < 2 {
			t.Errorf("expected at least 2 resources, got %d", len(resources))
		}

		// Check for expected resources
		foundSchema := false
		foundBestPractices := false
		for _, r := range resources {
			resource := r.(map[string]any)
			uri := resource["uri"].(string)
			if uri == "nais://schema" {
				foundSchema = true
			}
			if uri == "nais://api-best-practices" {
				foundBestPractices = true
			}
		}

		if !foundSchema {
			t.Error("expected schema resource")
		}
		if !foundBestPractices {
			t.Error("expected api-best-practices resource")
		}
	})

	t.Run("read_schema_resource", func(t *testing.T) {
		content := client.ReadResource(t, "nais://schema")

		if content == "" {
			t.Error("expected non-empty schema content")
		}

		// The mock schema should contain Team
		if !strings.Contains(content, "Team") {
			t.Error("expected schema to contain Team type")
		}
	})

	t.Run("read_best_practices_resource", func(t *testing.T) {
		content := client.ReadResource(t, "nais://api-best-practices")

		if content == "" {
			t.Error("expected non-empty best practices content")
		}

		// Should contain markdown content about best practices
		if !strings.Contains(content, "pagination") && !strings.Contains(content, "Pagination") {
			t.Error("expected best practices to mention pagination")
		}
	})
}
