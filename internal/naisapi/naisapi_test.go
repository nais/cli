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
	"golang.org/x/oauth2"
)

func newTestProxy(t *testing.T, upstream *httptest.Server) *httptest.Server {
	t.Helper()
	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream URL: %v", err)
	}
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
				if got := r.URL.RequestURI(); got != test.path {
					t.Errorf("upstream request URI = %q, want %q", got, test.path)
				}
				if r.Method != test.method {
					t.Errorf("upstream request method = %q, want %q", r.Method, test.method)
				}
				if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
					t.Errorf("upstream Authorization = %q, want %q", got, want)
				}
				if got, want := r.UserAgent(), "proxy-test (nais-cli)"; got != want {
					t.Errorf("upstream User-Agent = %q, want %q", got, want)
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read upstream request body: %v", err)
				}
				if got := string(body); got != test.body {
					t.Errorf("upstream request body = %q, want %q", got, test.body)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, responseBody)
			}))
			defer upstream.Close()
			proxy := newTestProxy(t, upstream)
			req, err := http.NewRequestWithContext(t.Context(), test.method, proxy.URL+test.path, strings.NewReader(test.body))
			if err != nil {
				t.Fatalf("create proxy request: %v", err)
			}
			req.Header.Set("User-Agent", "proxy-test")
			req.Header.Set("Authorization", "Bearer client-token")
			client := proxy.Client()
			client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("send proxy request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read proxy response body: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("proxy response status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if got, want := resp.Header.Get("Content-Type"), "application/json"; got != want {
				t.Errorf("proxy response Content-Type = %q, want %q", got, want)
			}
			if got := string(body); got != responseBody {
				t.Errorf("proxy response body = %q, want %q", got, responseBody)
			}
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
			if err != nil {
				t.Fatalf("send proxy request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read proxy response body: %v", err)
			}
			if resp.StatusCode != statusCode {
				t.Errorf("proxy response status = %d, want %d", resp.StatusCode, statusCode)
			}
			if got, want := resp.Header.Get("Content-Type"), "application/json"; got != want {
				t.Errorf("proxy response Content-Type = %q, want %q", got, want)
			}
			if got, want := resp.Header.Get("Retry-After"), "30"; got != want {
				t.Errorf("proxy response Retry-After = %q, want %q", got, want)
			}
			if got := string(body); got != responseBody {
				t.Errorf("proxy response body = %q, want %q", got, responseBody)
			}
		})
	}
}

func TestProxyUnavailableUpstream(t *testing.T) {
	upstream := httptest.NewTLSServer(http.NotFoundHandler())
	upstream.Close()
	proxy := newTestProxy(t, upstream)
	resp, err := proxy.Client().Get(proxy.URL + "/api/loki/api/v1/labels")
	if err != nil {
		t.Fatalf("send proxy request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("proxy response status = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
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
			if err != nil {
				t.Fatalf("send playground request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read playground response body: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("playground response status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if got, want := resp.Header.Get("Content-Type"), "text/html"; !strings.Contains(got, want) {
				t.Errorf("playground response Content-Type = %q, want it to contain %q", got, want)
			}
			html := string(body)
			for _, want := range []string{"Nais API playground", "/graphql"} {
				if !strings.Contains(html, want) {
					t.Errorf("playground response body does not contain %q: %q", want, html)
				}
			}
		})
	}
}

func TestProxyLokiTailWebSocket(t *testing.T) {
	path := "/api/loki/api/v1/tail?query=%7Bapp%3D%22test%22%7D&limit=10"
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RequestURI(); got != path {
			t.Errorf("upstream WebSocket request URI = %q, want %q", got, path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("upstream WebSocket request method = %q, want %q", r.Method, http.MethodGet)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Errorf("upstream WebSocket Authorization = %q, want %q", got, want)
		}
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade upstream WebSocket: %v", err)
			return
		}
		defer func() { _ = conn.Close() }()
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			t.Errorf("set upstream WebSocket read deadline: %v", err)
		}
		if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			t.Errorf("set upstream WebSocket write deadline: %v", err)
		}
		for range 2 {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				t.Errorf("read upstream WebSocket message: %v", err)
				return
			}
			if err := conn.WriteMessage(messageType, message); err != nil {
				t.Errorf("write upstream WebSocket message: %v", err)
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
	if err != nil {
		t.Fatalf("dial proxy WebSocket: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("proxy WebSocket response status = %d, want %d", resp.StatusCode, http.StatusSwitchingProtocols)
	}
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set proxy WebSocket read deadline: %v", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set proxy WebSocket write deadline: %v", err)
	}
	for _, message := range []string{`{"streams":[{"stream":{"app":"test"},"values":[["123","first"]]}]}`, `{"streams":[]}`} {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			t.Fatalf("write proxy WebSocket message: %v", err)
		}
		messageType, received, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read proxy WebSocket message: %v", err)
		}
		if messageType != websocket.TextMessage {
			t.Errorf("proxy WebSocket message type = %d, want %d", messageType, websocket.TextMessage)
		}
		if got := string(received); got != message {
			t.Errorf("proxy WebSocket message = %q, want %q", got, message)
		}
	}
}
