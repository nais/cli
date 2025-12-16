// Package resources provides MCP resource implementations for Nais data.
package resources

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
)

// RegisterResources registers all MCP resources with the server.
func RegisterResources(s *server.MCPServer, c client.Client, logger *slog.Logger) {
	ctx := &resourceContext{
		client: c,
		logger: logger,
	}

	// Register schema resource
	registerSchemaResource(s, ctx)

	// Register best practices resource
	registerBestPracticesResource(s, ctx)
}

// resourceContext holds shared dependencies for resource handlers.
type resourceContext struct {
	client client.Client
	logger *slog.Logger

	// Schema caching
	schemaOnce   sync.Once
	cachedSchema string
	schemaError  error
}

// getCachedSchema returns the cached schema.
// The schema is fetched once and cached for the lifetime of the resourceContext.
// Thread-safe using sync.Once.
func (ctx *resourceContext) getCachedSchema(reqCtx context.Context) (string, error) {
	ctx.schemaOnce.Do(func() {
		ctx.logger.Debug("Fetching and caching schema for resource")
		schema, err := ctx.client.GetSchema(reqCtx)
		if err != nil {
			ctx.schemaError = err
			ctx.logger.Error("Failed to fetch schema", "error", err)
			return
		}
		ctx.cachedSchema = schema
		ctx.logger.Debug("Schema cached successfully for resource")
	})

	return ctx.cachedSchema, ctx.schemaError
}

// registerSchemaResource registers the GraphQL schema resource.
func registerSchemaResource(s *server.MCPServer, ctx *resourceContext) {
	schemaResource := mcp.NewResource(
		"nais://schema",
		"GraphQL Schema",
		mcp.WithResourceDescription("The Nais GraphQL API schema"),
		mcp.WithMIMEType("text/plain"),
	)

	s.AddResource(schemaResource, func(reqCtx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		ctx.logger.Debug("Reading schema resource")

		schema, err := ctx.getCachedSchema(reqCtx)
		if err != nil {
			ctx.logger.Error("Failed to get schema", "error", err)
			return nil, err
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/plain",
				Text:     schema,
			},
		}, nil
	})
}

// registerBestPracticesResource registers the API best practices resource.
func registerBestPracticesResource(s *server.MCPServer, ctx *resourceContext) {
	bestPracticesResource := mcp.NewResource(
		"nais://api-best-practices",
		"API Best Practices",
		mcp.WithResourceDescription("Best practices and guidelines for using the Nais API, including pagination limits and query optimization"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.AddResource(bestPracticesResource, func(reqCtx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		ctx.logger.Debug("Reading best practices resource")

		content := `# Nais API Best Practices

## Pagination

When querying paginated connections (lists), always use reasonable page sizes:

- **Recommended page size**: 20-50 items
- **Maximum recommended**: 100 items
- **Never use**: 1000 or unlimited queries

### Example - Correct pagination:

` + "```" + `graphql
query TeamWorkloads($slug: Slug!, $first: Int!, $after: Cursor) {
  team(slug: $slug) {
    workloads(first: $first, after: $after) {
      nodes {
        name
        image { name tag }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}
` + "```" + `

Call with ` + "`first: 50`" + ` and use ` + "`endCursor`" + ` to fetch subsequent pages.

### Why pagination matters:

1. **Performance**: Large queries are slow and may timeout
2. **Resource usage**: Reduces load on the API server
3. **Reliability**: Smaller pages are less likely to fail

## Query Optimization

### Request only needed fields

Instead of requesting all fields, specify only what you need:

` + "```" + `graphql
# Good - only request needed fields
query {
  team(slug: "my-team") {
    applications(first: 50) {
      nodes {
        name
        state
      }
    }
  }
}
` + "```" + `

### Use filters when available

Most connections support filtering to reduce result size:

` + "```" + `graphql
query {
  team(slug: "my-team") {
    applications(first: 50, filter: { environments: ["prod"] }) {
      nodes { name state }
    }
  }
}
` + "```" + `

## Rate Limiting

- The API has rate limits to ensure fair usage
- If you receive rate limit errors, wait before retrying
- Use exponential backoff for retries

## Common Patterns

### Iterating over all teams

` + "```" + `graphql
query MyTeams {
  me {
    ... on User {
      teams(first: 50) {
        nodes {
          team { slug }
        }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}
` + "```" + `

### Getting workloads across environments

` + "```" + `graphql
query TeamWorkloads($slug: Slug!) {
  team(slug: $slug) {
    workloads(first: 50) {
      nodes {
        __typename
        name
        teamEnvironment {
          environment { name }
        }
        image { name tag }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
}
` + "```" + `
`

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/markdown",
				Text:     content,
			},
		}, nil
	})
}
