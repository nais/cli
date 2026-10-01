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
	states []Access
	calls  int
}

func (f *fakeAccessAPI) ActiveBranch(context.Context, string, string, string) (string, error) {
	return "main", nil
}

func (f *fakeAccessAPI) Create(context.Context, gql.CreatePostgresAccessInput) (string, error) {
	return "access-1", nil
}

func (f *fakeAccessAPI) Get(context.Context, string, string, string) (Access, error) {
	index := f.calls
	f.calls++
	if index >= len(f.states) {
		return Access{}, errors.New("unexpected poll")
	}
	return f.states[index], nil
}

func TestWaitForAccess(t *testing.T) {
	for _, tt := range []struct {
		name   string
		states []Access
		want   string
	}{
		{"ready", []Access{{State: gql.PostgresAccessStatePending}, {State: gql.PostgresAccessStateReady, Connection: &Connection{Username: "alice"}}}, ""},
		{"failed", []Access{{State: gql.PostgresAccessStateFailed, Message: "database unavailable"}}, "database unavailable"},
		{"expired", []Access{{State: gql.PostgresAccessStateExpired}}, "EXPIRED"},
		{"missing materials", []Access{{State: gql.PostgresAccessStateReady}}, "without connection materials"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAccessAPI{states: tt.states}
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
