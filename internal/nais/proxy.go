package nais

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func RunAPIProxy(ctx context.Context, addr string) error {
	tok, err := getUserToken(ctx)
	if err != nil {
		return err
	}

	// Setup reverse proxy to forward requests to the target server, but using a custom transport that authenticates the request
	target := &url.URL{
		Scheme: "https",
		Host:   tok.ConsoleHost,
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.Out.Header.Set("Host", tok.ConsoleHost)
			req.Out.Header.Set("Authorization", "Bearer "+tok.AccessToken)
			req.Out.Header.Set("user-agent", req.In.Header.Get("user-agent")+" (nais-api)")
		},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	fmt.Println("Forwarding requests from", addr, "to", target.String())
	// Start the server
	http.Handle("/", proxy)
	if err := http.ListenAndServe(addr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
