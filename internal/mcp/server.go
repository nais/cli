// Package mcp provides the MCP server implementation for Nais CLI.
package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
	"github.com/nais/cli/internal/mcp/resources"
	"github.com/nais/cli/internal/mcp/tools"
)

const (
	serverName    = "nais-mcp"
	serverVersion = "0.1.0"
)

// Server wraps the MCP server with Nais-specific configuration.
type Server struct {
	mcpServer   *server.MCPServer
	options     *Options
	rateLimiter *RateLimiter
	client      client.Client
}

// NewServer creates a new MCP server with the given options.
func NewServer(opts ...Option) (*Server, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Create the MCP server with capabilities
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false), // resources enabled, no subscriptions
		server.WithRecovery(),
	)

	// Create rate limiter
	rateLimiter := NewRateLimiter(options.RateLimit)

	// Determine which client to use
	var c client.Client
	if options.Client != nil {
		c = options.Client
	} else {
		c = client.NewLiveClient()
	}

	s := &Server{
		mcpServer:   mcpServer,
		options:     options,
		rateLimiter: rateLimiter,
		client:      c,
	}

	// Register tools and resources
	tools.RegisterTools(mcpServer, c, rateLimiter, options.Logger)
	resources.RegisterResources(mcpServer, c, options.Logger)

	return s, nil
}

// Serve starts the MCP server with the configured transport.
func (s *Server) Serve(ctx context.Context) error {
	switch s.options.Transport {
	case TransportStdio:
		return s.serveStdio()
	case TransportHTTP:
		return s.serveHTTP(ctx)
	case TransportSSE:
		return s.serveSSE(ctx)
	default:
		return fmt.Errorf("unknown transport: %s", s.options.Transport)
	}
}

// serveStdio starts the server with STDIO transport.
func (s *Server) serveStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// serveHTTP starts the server with HTTP transport.
func (s *Server) serveHTTP(ctx context.Context) error {
	httpServer := server.NewStreamableHTTPServer(s.mcpServer,
		server.WithStateLess(true),
	)
	s.options.Logger.Info("Starting HTTP server", "address", s.options.ListenAddr)
	return httpServer.Start(s.options.ListenAddr)
}

// serveSSE starts the server with SSE transport.
func (s *Server) serveSSE(ctx context.Context) error {
	sseServer := server.NewSSEServer(s.mcpServer)
	s.options.Logger.Info("Starting SSE server", "address", s.options.ListenAddr)
	return sseServer.Start(s.options.ListenAddr)
}

// MCPServer returns the underlying MCP server.
// This is useful for testing or advanced configuration.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}
