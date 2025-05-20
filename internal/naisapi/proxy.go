package naisapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func RunAPIProxy(ctx context.Context, addr string) error {
	secret, err := getUserSecret(ctx)
	if err != nil {
		return err
	}

	// Setup reverse proxy to forward requests to the target server, but using a custom transport that authenticates the request
	target := &url.URL{
		Scheme: "https",
		Host:   secret.ConsoleHost,
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.Out.Header.Set("Host", secret.ConsoleHost)
			req.Out.Header.Set("Authorization", "Bearer "+secret.AccessToken)
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
