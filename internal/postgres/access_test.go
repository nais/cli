package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi/gql"
)

type fakeAccessAPI struct {
	states          []Access
	connection      Connection
	connectionErr   error
	calls           int
	connectionCalls int
}

func (f *fakeAccessAPI) Create(context.Context, gql.CreatePostgresAccessInput) (string, error) {
	return "access-1", nil
}

func TestCreateAccessOmitsBranchOnTheWire(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			OperationName string `json:"operationName"`
			Variables     struct {
				Input map[string]json.RawMessage `json:"input"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.OperationName != "CreatePostgresAccess" {
			t.Errorf("unexpected branch lookup: %s", request.OperationName)
		}
		if _, present := request.Variables.Input["branch"]; present {
			t.Error("branchless access still sends the deprecated branch field")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"createPostgresAccess":{"name":"access-1"}}}`))
	}))
	defer server.Close()
	api := graphqlAccessAPI{client: graphql.NewClient(server.URL, server.Client())}
	name, err := api.Create(t.Context(), gql.CreatePostgresAccessInput{Postgres: "orders", TeamSlug: "team", EnvironmentName: "dev", AccessLevel: gql.PostgresAccessLevelRead, Reason: "Investigating issue"})
	if err != nil || name != "access-1" {
		t.Fatalf("branchless access = %q, error = %v", name, err)
	}
}

func (f *fakeAccessAPI) Status(context.Context, string, string, string) (Access, error) {
	index := f.calls
	f.calls++
	if index >= len(f.states) {
		return Access{}, errors.New("unexpected poll")
	}
	return f.states[index], nil
}

func (f *fakeAccessAPI) Connection(context.Context, string, string, string) (Connection, error) {
	f.connectionCalls++
	return f.connection, f.connectionErr
}

func TestWaitForAccess(t *testing.T) {
	for _, tt := range []struct {
		name          string
		states        []Access
		connectionErr error
		want          string
		wantConnCalls int
	}{
		{"ready after pending", []Access{{State: gql.PostgresAccessStatePending}, {State: gql.PostgresAccessStatePending}, {State: gql.PostgresAccessStateReady}}, nil, "", 1},
		{"failed", []Access{{State: gql.PostgresAccessStateFailed, Message: "database unavailable"}}, nil, "database unavailable", 0},
		{"expired", []Access{{State: gql.PostgresAccessStateExpired}}, nil, "EXPIRED", 0},
		{"ready but connection fails", []Access{{State: gql.PostgresAccessStateReady}}, errors.New("secret missing"), "secret missing", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAccessAPI{states: tt.states, connection: Connection{Username: "alice"}, connectionErr: tt.connectionErr}
			got, err := waitForAccess(context.Background(), fake, "team", "dev", "access-1", time.Millisecond)
			if tt.want == "" {
				if err != nil || got.Username != "alice" {
					t.Fatalf("got %+v, err %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
			if fake.calls != len(tt.states) {
				t.Fatalf("polled %d times, want %d", fake.calls, len(tt.states))
			}
			// Credentials are only requested once the access is ready.
			if fake.connectionCalls != tt.wantConnCalls {
				t.Fatalf("connection fetched %d times, want %d", fake.connectionCalls, tt.wantConnCalls)
			}
		})
	}
}

func TestWaitCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fake := &fakeAccessAPI{states: []Access{{State: gql.PostgresAccessStatePending}}}
	_, err := waitForAccess(ctx, fake, "team", "dev", "access-1", time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
