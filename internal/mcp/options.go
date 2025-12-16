// Package mcp provides the MCP server implementation for Nais CLI.
package mcp

import (
	"io"
	"log/slog"

	"github.com/nais/cli/internal/mcp/client"
)

// Transport defines the transport type for the MCP server.
type Transport string

const (
	// TransportStdio uses standard input/output for communication.
	TransportStdio Transport = "stdio"
	// TransportHTTP uses HTTP for communication.
	TransportHTTP Transport = "http"
	// TransportSSE uses Server-Sent Events for communication.
	TransportSSE Transport = "sse"
)

// Options holds the configuration for the MCP server.
type Options struct {
	// Transport specifies the transport type (stdio, http, sse).
	Transport Transport

	// ListenAddr is the address to listen on for HTTP/SSE transports.
	ListenAddr string

	// RateLimit is the maximum requests per minute (0 = unlimited).
	RateLimit int

	// Logger is the logger for MCP operations.
	Logger *slog.Logger

	// LogOutput is where logs are written (defaults to stderr).
	LogOutput io.Writer

	// Client is the GraphQL client to use. If nil, uses the live client.
	Client client.Client
}

// Option is a functional option for configuring the MCP server.
type Option func(*Options)

// DefaultOptions returns the default options for the MCP server.
func DefaultOptions() *Options {
	return &Options{
		Transport:  TransportStdio,
		ListenAddr: ":8080",
		RateLimit:  10,
		Logger:     slog.Default(),
	}
}

// WithTransport sets the transport type.
func WithTransport(t Transport) Option {
	return func(o *Options) {
		o.Transport = t
	}
}

// WithListenAddr sets the listen address for HTTP/SSE transports.
func WithListenAddr(addr string) Option {
	return func(o *Options) {
		o.ListenAddr = addr
	}
}

// WithRateLimit sets the rate limit (requests per minute).
func WithRateLimit(limit int) Option {
	return func(o *Options) {
		o.RateLimit = limit
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithLogOutput sets the log output destination.
func WithLogOutput(w io.Writer) Option {
	return func(o *Options) {
		o.LogOutput = w
	}
}

// WithClient sets the GraphQL client.
func WithClient(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}
