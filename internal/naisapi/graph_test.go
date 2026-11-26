package naisapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, CheckAuthentication(ctx))
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
			require.NoError(t, err)
			var data struct{ Email string }
			err = client.MakeRequest(tt.ctx, &graphql.Request{Query: "{ email }"}, &graphql.Response{Data: &data})
			require.NoError(t, err)
			require.Equal(t, tt.want, data.Email)
		})
	}
}
