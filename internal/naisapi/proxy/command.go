package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/nais/cli/internal/naisapi"
)

type Flags struct {
	*naisapi.Flags
	ListenAddr string
}

func Run(ctx context.Context, flags *Flags) error {
	user, err := naisapi.GetAuthenticatedUser(ctx)
	if err != nil {
		return err
	}

	// Setup reverse proxy to forward requests to the target server, but using a custom transport that authenticates the request
	target := &url.URL{
		Scheme: "https",
		Host:   user.ConsoleHost,
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.Out.Header.Set("user-agent", req.In.Header.Get("user-agent")+" (nais-cli)")
		},
		Transport: user.RoundTripper(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}),
	}

	fmt.Println("Forwarding requests from", flags.ListenAddr, "to", target.String())
	// Start the server
	http.Handle("/", proxy)
	if err := http.ListenAndServe(flags.ListenAddr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
