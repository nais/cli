package flag

import (
	"context"

	"github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/naistrix"
)

type MCP struct {
	*flag.Alpha
}

// Transport represents the transport type for the MCP server.
type Transport string

var _ naistrix.FlagAutoCompleter = (*Transport)(nil)

func (t *Transport) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"stdio", "http", "sse"}, "Available transport types."
}

type Serve struct {
	*MCP
	Transport  Transport `name:"transport" short:"t" usage:"Transport type (stdio, http, sse)."`
	ListenAddr string    `name:"listen" short:"l" usage:"Address to listen on (for http/sse transports)."`
	RateLimit  int       `name:"rate-limit" short:"r" usage:"Maximum requests per minute (0 = unlimited)."`
	LogFile    string    `name:"log-file" usage:"Write logs to file instead of stderr."`
}
