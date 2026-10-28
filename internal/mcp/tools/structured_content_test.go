package tools

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

type allowAll struct{}

func (allowAll) Allow() bool { return true }

// MCP requires structuredContent to be a JSON object, and outputSchema to
// describe one.
func TestStructuredContentIsObject(t *testing.T) {
	s := server.NewMCPServer("test", "0.0.0", server.WithToolCapabilities(false))
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	RegisterTools(s, &mockClientWithCounter{schema: `
type Query { me: String, team(slug: String!): String }
type Mutation { createTeam(slug: String!): String }
`}, allowAll{}, logger)

	call := func(method string, params any) map[string]any {
		t.Helper()
		req, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
		raw, _ := json.Marshal(s.HandleMessage(context.Background(), req))
		var resp struct {
			Result map[string]any `json:"result"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil || resp.Result == nil {
			t.Fatalf("%s: unexpected response: %s", method, raw)
		}
		return resp.Result
	}

	for _, tool := range call("tools/list", nil)["tools"].([]any) {
		tool := tool.(map[string]any)
		schema, ok := tool["outputSchema"].(map[string]any)
		if !ok {
			continue
		}
		if schema["type"] != "object" {
			t.Errorf("%s: outputSchema type = %v, want object", tool["name"], schema["type"])
		}
		if _, ok := schema["items"]; ok {
			t.Errorf("%s: outputSchema has items, but describes an object", tool["name"])
		}
	}

	for name, want := range map[string]struct {
		key   string
		count int
	}{
		"schema_list_queries":   {"queries", 4}, // me, team, __schema, __type,
		"schema_list_mutations": {"mutations", 1},
	} {
		result := call("tools/call", map[string]any{"name": name, "arguments": map[string]any{}})
		sc, ok := result["structuredContent"].(map[string]any)
		if !ok {
			t.Errorf("%s: structuredContent is not an object: %v", name, result)
			continue
		}
		list, ok := sc[want.key].([]any)
		if !ok || len(list) != want.count {
			t.Errorf("%s: structuredContent[%q] = %v, want %d items", name, want.key, sc[want.key], want.count)
		}
	}
}
