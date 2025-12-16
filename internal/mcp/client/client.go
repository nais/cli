// Package client provides the GraphQL client interface for MCP operations.
package client

import (
	"context"

	"github.com/nais/cli/internal/naisapi/gql"
)

// Client defines the interface for GraphQL operations used by MCP tools.
// This interface allows for both live API clients and mock clients for testing.
type Client interface {
	// User operations
	GetCurrentUser(ctx context.Context) (*User, error)
	GetUserTeams(ctx context.Context) ([]gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMember, error)

	// Schema operations
	GetSchema(ctx context.Context) (string, error)

	// Console URL operations
	GetConsoleURL(ctx context.Context) (string, error)
}

// User represents the authenticated user.
type User struct {
	Name    string
	Email   string
	IsAdmin bool
}
