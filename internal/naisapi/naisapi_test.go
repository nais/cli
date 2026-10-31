package naisapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestProxy(t *testing.T, upstream *httptest.Server) *httptest.Server {
	t.Helper()
	target, err := url.Parse(upstream.URL)
	require.NoError(t, err)
	transport := &oauth2.Transport{
		Base:   upstream.Client().Transport,
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}),
	}
	proxy := httptest.NewServer(newProxyHandler(target, transport))
	t.Cleanup(proxy.Close)
	return proxy
}

func TestProxyRequests(t *testing.T) {
	for _, test := range []struct {
		name   string
		path   string
		method string
		body   string
	}{
		{
			name:   "Loki root without redirect",
			path:   "/api/loki?query=up",
			method: http.MethodGet,
		},
		{
			name:   "Loki labels",
			path:   "/api/loki/api/v1/labels?start=123&end=456",
			method: http.MethodGet,
		},
		{
			name:   "Loki query range",
			path:   "/api/loki/api/v1/query_range?query=%7Bapp%3D%22test%22%7D+%7C%3D+%22hello%20world%22&start=123&end=456&limit=10&direction=backward&step=2&extra=one&extra=two",
			method: http.MethodGet,
		},
		{
			name:   "Loki POST body",
			path:   "/api/loki/api/v1/query_range",
			method: http.MethodPost,
			body:   "query=%7Bapp%3D%22test%22%7D&start=123&end=456",
		},
		{
			name:   "Loki escaped path",
			path:   "/api/loki/api/v1/label/app%2Fname/values?start=123",
			method: http.MethodGet,
		},
		{
			name:   "GraphQL",
			path:   "/graphql?operationName=Me",
			method: http.MethodPost,
			body:   `{"query":"{ me { __typename } }"}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			responseBody := `{"status":"success","data":[]}`
			upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.path, r.URL.RequestURI())
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "proxy-test (nais-cli)", r.UserAgent())
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Equal(t, test.body, string(body))
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, responseBody)
			}))
			defer upstream.Close()
			proxy := newTestProxy(t, upstream)
			req, err := http.NewRequestWithContext(t.Context(), test.method, proxy.URL+test.path, strings.NewReader(test.body))
			require.NoError(t, err)
			req.Header.Set("User-Agent", "proxy-test")
			req.Header.Set("Authorization", "Bearer client-token")
			client := proxy.Client()
			client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			assert.Equal(t, responseBody, string(body))
		})
	}
}

func TestProxyUpstreamErrors(t *testing.T) {
	for _, statusCode := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			responseBody := `{"status":"error","error":"upstream failure"}`
			upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "30")
				w.WriteHeader(statusCode)
				_, _ = io.WriteString(w, responseBody)
			}))
			defer upstream.Close()
			proxy := newTestProxy(t, upstream)
			resp, err := proxy.Client().Get(proxy.URL + "/api/loki/api/v1/query_range")
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, statusCode, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			assert.Equal(t, "30", resp.Header.Get("Retry-After"))
			assert.Equal(t, responseBody, string(body))
		})
	}
}

func TestProxyUnavailableUpstream(t *testing.T) {
	upstream := httptest.NewTLSServer(http.NotFoundHandler())
	upstream.Close()
	proxy := newTestProxy(t, upstream)
	resp, err := proxy.Client().Get(proxy.URL + "/api/loki/api/v1/labels")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestProxyPlayground(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("Playground request reached upstream")
	}))
	defer upstream.Close()
	proxy := newTestProxy(t, upstream)
	for _, path := range []string{"/", "/other", "/api/loki-other"} {
		t.Run(path, func(t *testing.T) {
			resp, err := proxy.Client().Get(proxy.URL + path)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
			assert.Contains(t, string(body), "Nais API playground")
			assert.Contains(t, string(body), "/graphql")
		})
	}
}

func TestProxyLokiTailWebSocket(t *testing.T) {
	path := "/api/loki/api/v1/tail?query=%7Bapp%3D%22test%22%7D&limit=10"
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.RequestURI())
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if !assert.NoError(t, err) {
			return
		}
		defer func() { _ = conn.Close() }()
		assert.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
		assert.NoError(t, conn.SetWriteDeadline(time.Now().Add(5*time.Second)))
		for range 2 {
			messageType, message, err := conn.ReadMessage()
			if !assert.NoError(t, err) {
				return
			}
			if !assert.NoError(t, conn.WriteMessage(messageType, message)) {
				return
			}
		}
	}))
	defer upstream.Close()
	proxy := newTestProxy(t, upstream)
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, resp, err := dialer.DialContext(t.Context(), "ws"+strings.TrimPrefix(proxy.URL, "http")+path, nil)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	require.NoError(t, conn.SetWriteDeadline(time.Now().Add(5*time.Second)))
	for _, message := range []string{`{"streams":[{"stream":{"app":"test"},"values":[["123","first"]]}]}`, `{"streams":[]}`} {
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(message)))
		messageType, received, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, websocket.TextMessage, messageType)
		assert.Equal(t, message, string(received))
	}
}
