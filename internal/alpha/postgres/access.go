package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// Access contains the brokered connection materials. Do not log this value.
type Access struct {
	State      gql.PostgresAccessState
	Message    string
	Connection *Connection
}

type Connection struct {
	Username, Password, CACertificate, ServerName, RelayEndpoint, RelayAccess, RelayToken string
}

type AccessAPI interface {
	ActiveBranch(context.Context, string, string, string) (string, error)
	Create(context.Context, gql.CreatePostgresAccessInput) (string, error)
	Get(context.Context, string, string, string) (Access, error)
}

type graphqlAccessAPI struct{ client graphql.Client }

func NewAPI(ctx context.Context) (AccessAPI, error) {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlAccessAPI{client}, nil
}

func (a graphqlAccessAPI) ActiveBranch(ctx context.Context, team, environment, name string) (string, error) {
	_ = `# @genqlient
 query GetActivePostgresBranchAlpha($team: Slug!, $environment: String!, $postgres: String!) {
  team(slug: $team) { environment(name: $environment) { postgres(name: $postgres) { activeBranch { name } } } }
 }
 `
	result, err := gql.GetActivePostgresBranchAlpha(ctx, a.client, team, environment, name)
	if err != nil {
		return "", err
	}
	if result.Team.Environment.Postgres.ActiveBranch == nil {
		return "", fmt.Errorf("postgres %q has no active branch; specify --branch", name)
	}
	return result.Team.Environment.Postgres.ActiveBranch.Name, nil
}

func (a graphqlAccessAPI) Create(ctx context.Context, input gql.CreatePostgresAccessInput) (string, error) {
	_ = `# @genqlient
 mutation CreatePostgresAccessAlpha($input: CreatePostgresAccessInput!) {
  createPostgresAccess(input: $input) { name }
 }
 `
	result, err := gql.CreatePostgresAccessAlpha(ctx, a.client, input)
	if err != nil {
		return "", err
	}
	return result.CreatePostgresAccess.Name, nil
}

func (a graphqlAccessAPI) Get(ctx context.Context, team, environment, name string) (Access, error) {
	_ = `# @genqlient
 query GetPostgresAccessAlpha($team: Slug!, $environment: String!, $name: String!) {
  team(slug: $team) { environment(name: $environment) { postgresAccess(name: $name) {
   state message connection { username password caCertificate serverName relayEndpoint relayAccess relayToken }
  } } }
 }
 `
	result, err := gql.GetPostgresAccessAlpha(ctx, a.client, team, environment, name)
	if err != nil {
		return Access{}, err
	}
	got := result.Team.Environment.PostgresAccess
	access := Access{State: got.State}
	if got.Message != nil {
		access.Message = *got.Message
	}
	if got.Connection != nil {
		c := got.Connection
		access.Connection = &Connection{c.Username, c.Password, c.CaCertificate, c.ServerName, c.RelayEndpoint, c.RelayAccess, c.RelayToken}
	}
	return access, nil
}

func waitForAccess(ctx context.Context, api AccessAPI, team, environment, name string, interval time.Duration) (Connection, error) {
	for {
		access, err := api.Get(ctx, team, environment, name)
		if err != nil {
			return Connection{}, fmt.Errorf("retrieve postgres access: %w", err)
		}
		switch access.State {
		case gql.PostgresAccessStateReady:
			if access.Connection == nil {
				return Connection{}, fmt.Errorf("postgres access %q is ready without connection materials", name)
			}
			return *access.Connection, nil
		case gql.PostgresAccessStateFailed, gql.PostgresAccessStateExpired:
			return Connection{}, fmt.Errorf("postgres access %q is %s: %s", name, access.State, access.Message)
		case gql.PostgresAccessStatePending:
		default:
			return Connection{}, fmt.Errorf("postgres access %q has unknown state %q", name, access.State)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Connection{}, fmt.Errorf("waiting for postgres access %q: %w", name, ctx.Err())
		case <-timer.C:
		}
	}
}

func CreateAndWait(ctx context.Context, api AccessAPI, input gql.CreatePostgresAccessInput) (Connection, error) {
	// Bound creation and polling together; a stalled API must not hang the command.
	setupCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	name, err := api.Create(setupCtx, input)
	if err != nil {
		return Connection{}, fmt.Errorf("create postgres access: %w", err)
	}
	connection, err := waitForAccess(setupCtx, api, input.TeamSlug, input.EnvironmentName, name, time.Second)
	if err != nil {
		return Connection{}, fmt.Errorf("access %q was created but is not ready (it expires after its requested TTL): %w", name, err)
	}
	return connection, nil
}
