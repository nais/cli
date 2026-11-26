package naisapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
)

func TestGraphqlClientInvocationAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"email":"`+r.Header.Get("X-User-Email")+`"}}`)
	}))
	defer server.Close()
	t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
	t.Setenv("NAIS_API_LOCAL_EMAIL", "checked@example.com")

	ctx := WithInvocationAuth(t.Context())
	if err := CheckAuthentication(ctx); err != nil {
		t.Fatalf("check authentication: %v", err)
	}
	// A fresh lookup would now produce a different user.
	t.Setenv("NAIS_API_LOCAL_EMAIL", "fallback@example.com")

	for _, tt := range []struct {
		name string
		ctx  context.Context
		want string
	}{
		{"reuse checked user", ctx, "checked@example.com"},
		{"another invocation without a check", WithInvocationAuth(t.Context()), "fallback@example.com"},
		{"without invocation context", t.Context(), "fallback@example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, err := GraphqlClient(tt.ctx)
			if err != nil {
				t.Fatalf("create GraphQL client: %v", err)
			}
			var data struct{ Email string }
			if err := client.MakeRequest(tt.ctx, &graphql.Request{Query: "{ email }"}, &graphql.Response{Data: &data}); err != nil {
				t.Fatalf("make GraphQL request: %v", err)
			}
			if data.Email != tt.want {
				t.Fatalf("email = %q, want %q", data.Email, tt.want)
			}
		})
	}
}
