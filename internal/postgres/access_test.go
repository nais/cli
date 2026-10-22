package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
)

type fakeAccessAPI struct {
	states          []Access
	connection      Connection
	connectionErr   error
	calls           int
	connectionCalls int
}

func (f *fakeAccessAPI) ActiveBranch(context.Context, string, string, string) (string, error) {
	return "main", nil
}

func (f *fakeAccessAPI) Create(context.Context, gql.CreatePostgresAccessInput) (string, error) {
	return "access-1", nil
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
