// Package relay forwards local TCP connections over authenticated HTTP/3 CONNECT streams.
package relay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/quic-go/quic-go/http3"
)

// Tunnel is a brokered relay endpoint and owner-only proof. Never log its token.
type Tunnel struct{ Endpoint, Access, Token string }

func (t Tunnel) request(ctx context.Context, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(t.Endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.User != nil || u.Fragment != "" {
		return nil, fmt.Errorf("invalid relay endpoint")
	}
	if !strings.Contains(t.Access, "/") || t.Token == "" {
		return nil, fmt.Errorf("invalid relay access credentials")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodConnect, t.Endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Host = u.Host // CONNECT authority is the relay, not the database target.
	req.Header.Set("Authorization", "Bearer "+t.Token)
	req.Header.Set("Relay-Access", t.Access)
	return req, nil
}

// Serve forwards each TCP connection to the relay until ctx is cancelled.
func Serve(ctx context.Context, listener net.Listener, tunnel Tunnel) error {
	transport := &http3.Transport{}
	defer func() { _ = transport.Close() }()
	defer func() { _ = listener.Close() }()
	stop := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer stop()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept local connection: %w", err)
		}
		go func() {
			defer func() { _ = conn.Close() }()
			if err := forward(ctx, transport, tunnel, conn); err != nil && ctx.Err() == nil {
				// An individual connection must not terminate other local clients.
				// Callers can retry; no credentials are included in the error.
				fmt.Fprintf(os.Stderr, "postgres relay connection failed: %v\n", err)
			}
		}()
	}
}

func forward(ctx context.Context, transport *http3.Transport, tunnel Tunnel, local net.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(ctx, func() { _ = local.Close() })
	defer stop()
	reader, writer := io.Pipe()
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()
	req, err := tunnel.request(ctx, reader)
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(writer, local)
		_ = writer.CloseWithError(err) // Send HTTP request FIN when TCP input is closed.
		done <- err
	}()
	response, err := transport.RoundTrip(req)
	if err != nil {
		_ = local.Close()
		_ = reader.CloseWithError(err)
		<-done
		return fmt.Errorf("relay CONNECT: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		_ = local.Close()
		_ = reader.Close()
		<-done
		return fmt.Errorf("relay CONNECT returned HTTP %d", response.StatusCode)
	}
	_, downloadErr := io.Copy(local, response.Body)
	if tcp, ok := local.(interface{ CloseWrite() error }); ok && downloadErr == nil {
		downloadErr = tcp.CloseWrite()
	}
	if downloadErr != nil {
		_ = local.Close()
		_ = reader.CloseWithError(downloadErr)
	}
	uploadErr := <-done
	if downloadErr != nil {
		return downloadErr
	}
	if uploadErr != nil && !errors.Is(uploadErr, net.ErrClosed) {
		return uploadErr
	}
	return nil
}
