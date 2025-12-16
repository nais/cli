package command

import (
	"context"
	"io"
	"log/slog"
	"os"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/mcp"
	"github.com/nais/cli/internal/mcp/command/flag"
	"github.com/nais/naistrix"
)

func MCP(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.MCP{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "mcp",
		Title:       "Model Context Protocol server for LLM-assisted workflows.",
		Description: "Start an MCP server that exposes Nais API operations as tools for LLMs. Supports multiple transports (stdio, http, sse).",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			serveCommand(flags),
		},
	}
}

func serveCommand(parentFlags *flag.MCP) *naistrix.Command {
	flags := &flag.Serve{
		MCP:        parentFlags,
		Transport:  "stdio",
		ListenAddr: ":8080",
		RateLimit:  25,
	}

	return &naistrix.Command{
		Name:  "serve",
		Title: "Start the MCP server.",
		Description: `Start the MCP server with the specified transport.

Examples:
  # Start with stdio transport (default, for Claude Desktop)
  nais alpha mcp serve

  # Start with HTTP transport on a specific port
  nais alpha mcp serve --transport http --listen :8080

  # Start with SSE transport
  nais alpha mcp serve --transport sse --listen :8080

  # Restrict to specific teams
  nais alpha mcp serve --team my-team --team other-team

  # Set rate limit
  nais alpha mcp serve --rate-limit 20`,
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return runServe(ctx, flags, out)
		},
	}
}

func runServe(ctx context.Context, flags *flag.Serve, out *naistrix.OutputWriter) error {
	// Configure logging
	var logOutput io.Writer = os.Stderr
	if flags.LogFile != "" {
		f, err := os.OpenFile(flags.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return err
		}
		defer f.Close()
		logOutput = f
	}

	// Map verbose flag to log level:
	// -v (verbose) = Info, -vv (debug) = Debug, -vvv (trace) = Debug with more detail
	logLevel := slog.LevelWarn
	if flags.IsVerbose() {
		logLevel = slog.LevelInfo
	}
	if flags.IsDebug() {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(logOutput, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Build server options
	opts := []mcp.Option{
		mcp.WithTransport(mcp.Transport(flags.Transport)),
		mcp.WithListenAddr(flags.ListenAddr),
		mcp.WithRateLimit(flags.RateLimit),
		mcp.WithLogger(logger),
		mcp.WithLogOutput(logOutput),
	}

	// Create and start the server
	server, err := mcp.NewServer(opts...)
	if err != nil {
		return err
	}

	// Log startup info (only for non-stdio transports to avoid polluting the protocol)
	if flags.Transport != "stdio" {
		out.Println("Starting MCP server with transport:", string(flags.Transport))
		if flags.Transport == "http" || flags.Transport == "sse" {
			out.Println("Listening on:", flags.ListenAddr)
		}
	}

	return server.Serve(ctx)
}
