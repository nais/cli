// Package tools provides MCP tool implementations for Nais operations.
package tools

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
)

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	Allow() bool
}

// RegisterTools registers all MCP tools with the server.
//
// The tools are organized into two categories:
// 1. Schema exploration tools - for discovering the GraphQL API structure
// 2. GraphQL execution tools - for executing queries against the Nais API
//
// This approach allows LLMs to dynamically explore the schema and construct
// queries based on user needs, rather than relying on a fixed set of specialized tools.
func RegisterTools(s *server.MCPServer, c client.Client, rateLimiter RateLimiter, logger *slog.Logger) {
	logger.Debug("Starting tool registration")

	ctx := &toolContext{
		client:      c,
		rateLimiter: rateLimiter,
		logger:      logger,
	}

	// Register schema exploration tools (needed for LLM to understand the API)
	logger.Debug("Registering schema tools")
	registerSchemaTools(s, ctx)

	// Register GraphQL execution tools (for dynamic query execution)
	logger.Debug("Registering GraphQL tools")
	registerGraphQLTools(s, ctx)

	logger.Debug("All tools registered successfully")
}

// toolContext holds shared dependencies for tool handlers.
type toolContext struct {
	client      client.Client
	rateLimiter RateLimiter
	logger      *slog.Logger

	// Schema caching
	schemaOnce   sync.Once
	cachedSchema string
	schemaError  error
}

// getConsoleBaseURL returns the base console URL for generating links.
// Returns an empty string if the console URL cannot be determined.
func (t *toolContext) getConsoleBaseURL(reqCtx context.Context) string {
	baseURL, err := t.client.GetConsoleURL(reqCtx)
	if err != nil {
		t.logger.Debug("Failed to get console URL", "error", err)
		return ""
	}
	return baseURL
}

// getCachedSchema returns the cached and repaired schema.
// The schema is fetched once and cached for the lifetime of the toolContext.
// Thread-safe using sync.Once.
func (t *toolContext) getCachedSchema(reqCtx context.Context) (string, error) {
	t.schemaOnce.Do(func() {
		t.logger.Debug("Fetching and caching schema")
		rawSchema, err := t.client.GetSchema(reqCtx)
		if err != nil {
			t.schemaError = err
			t.logger.Error("Failed to fetch schema", "error", err)
			return
		}

		// Repair the schema by removing built-in scalar redeclarations
		t.cachedSchema = removeBuiltinScalars(rawSchema)
		t.logger.Debug("Schema cached successfully")
	})

	return t.cachedSchema, t.schemaError
}
