package tools

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"slices"
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

	vulnerabilityToolFound := false
	for _, tool := range call("tools/list", nil)["tools"].([]any) {
		tool := tool.(map[string]any)
		if tool["name"] == "get_workload_vulnerabilities" {
			vulnerabilityToolFound = true
			schema, ok := tool["inputSchema"].(map[string]any)
			if !ok {
				t.Fatalf("vulnerabilities: missing input schema: %v", tool)
			}
			properties, ok := schema["properties"].(map[string]any)
			if !ok || schema["type"] != "object" {
				t.Fatalf("vulnerabilities: missing input properties: %v", schema)
			}
			for name, want := range map[string]string{
				"team": "string", "workload": "string", "environment": "string",
				"workload_type": "string", "first": "integer", "after": "string",
				"expected_snapshot": "string",
			} {
				property, ok := properties[name].(map[string]any)
				types, nullable := property["type"].([]any)
				if !ok || property["type"] != want && (!nullable || len(types) != 2 ||
					!slices.Contains(types, any(want)) || !slices.Contains(types, any("null"))) {
					t.Fatalf("vulnerabilities: input %q = %v, want %s", name, properties[name], want)
				}
			}
			required, ok := schema["required"].([]any)
			if !ok || len(required) != 4 {
				t.Fatalf("vulnerabilities: required inputs = %v", schema["required"])
			}
			for _, name := range []string{"team", "workload", "environment", "workload_type"} {
				if !slices.Contains(required, any(name)) {
					t.Errorf("vulnerabilities: %q is not required", name)
				}
			}
			workloadType := properties["workload_type"].(map[string]any)
			enum, ok := workloadType["enum"].([]any)
			if !ok || len(enum) != 2 || !slices.Contains(enum, any("application")) || !slices.Contains(enum, any("job")) {
				t.Errorf("vulnerabilities: workload_type enum = %v", workloadType["enum"])
			}
		}
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
	if !vulnerabilityToolFound {
		t.Fatal("get_workload_vulnerabilities is not registered")
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

	result := call("tools/call", map[string]any{
		"name": "get_workload_vulnerabilities",
		"arguments": map[string]any{
			"team": "team", "workload": "app", "environment": "prod-gcp",
			"workload_type": "application", "first": 1,
		},
	})
	sc, ok := result["structuredContent"].(map[string]any)
	if !ok || result["isError"] == true {
		t.Fatalf("vulnerabilities: missing structured output: %v", result)
	}
	if sc["workload"].(map[string]any)["name"] != "app" || sc["returned_count"] != float64(1) {
		t.Fatalf("vulnerabilities: wrong scope/count: %v", sc)
	}
	info := sc["page_info"].(map[string]any)
	if info["has_next_page"] != true || info["requested_after"] != nil {
		t.Fatalf("vulnerabilities: invalid first page: %v", info)
	}
	next := call("tools/call", map[string]any{
		"name": "get_workload_vulnerabilities",
		"arguments": map[string]any{
			"team": "team", "workload": "app", "environment": "prod-gcp",
			"workload_type": "application", "first": 1,
			"after": info["end_cursor"], "expected_snapshot": sc["snapshot_identity"],
		},
	})
	nextSC, ok := next["structuredContent"].(map[string]any)
	if !ok || next["isError"] == true || nextSC["snapshot_identity"] != sc["snapshot_identity"] {
		t.Fatalf("vulnerabilities: invalid continuation: %v", next)
	}
	if nextSC["page_info"].(map[string]any)["requested_after"] != info["end_cursor"] {
		t.Fatalf("vulnerabilities: continuation cursor not retained: %v", nextSC)
	}
}
