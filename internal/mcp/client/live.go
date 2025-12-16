//go:build !mock

// Package client provides the GraphQL client interface for MCP operations.
package client

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// LiveClient implements the Client interface using the real Nais API.
type LiveClient struct{}

// NewLiveClient creates a new live client.
func NewLiveClient() *LiveClient {
	return &LiveClient{}
}

// GetCurrentUser returns the current authenticated user.
func (c *LiveClient) GetCurrentUser(ctx context.Context) (*User, error) {
	user, err := naisapi.GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	isAdmin := naisapi.IsConsoleAdmin(ctx)

	return &User{
		Name:    user.Email(), // AuthenticatedUser interface only has Email(), not Name()
		Email:   user.Email(),
		IsAdmin: isAdmin,
	}, nil
}

// GetUserTeams returns the teams the current user is a member of.
func (c *LiveClient) GetUserTeams(ctx context.Context) ([]gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMember, error) {
	return naisapi.GetUserTeams(ctx)
}

// GetSchema returns the GraphQL schema.
func (c *LiveClient) GetSchema(ctx context.Context) (string, error) {
	return naisapi.PullSchema(ctx, nil)
}

// GetConsoleURL returns the base console URL for the current tenant.
func (c *LiveClient) GetConsoleURL(ctx context.Context) (string, error) {
	user, err := naisapi.GetAuthenticatedUser(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s", user.ConsoleHost()), nil
}
