package relay

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

// 32 bytes, unpadded base64url, as issued by the relay contract.
var (
	proofToken  = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	deniedToken = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32))
)

func TestConnectStreamsAfterHalfClose(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"}, DNSNames: []string{"localhost"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"}, DNSNames: []string{"localhost"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert.Leaf)
	packet, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer "+deniedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodConnect || r.Host != "localhost:"+strings.Split(packet.LocalAddr().String(), ":")[1] || r.Header.Get("Authorization") != "Bearer "+proofToken || r.Header.Get("Relay-Access") != "team/access" {
			t.Errorf("unexpected CONNECT: method=%s host=%s auth=%s access=%s", r.Method, r.Host, r.Header.Get("Authorization"), r.Header.Get("Relay-Access"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = http.NewResponseController(w).Flush()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		_, _ = w.Write([]byte("reply:" + string(data)))
	})
	server := &http3.Server{Handler: handler, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}
	serverDone := make(chan struct{})
	go func() { defer close(serverDone); _ = server.Serve(packet) }()
	defer func() { _ = server.Close(); <-serverDone; _ = packet.Close() }()
	transport := &http3.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, ServerName: "localhost"}}
	defer func() { _ = transport.Close() }()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	_ = client.SetDeadline(time.Now().Add(5 * time.Second))
	local, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = local.Close() }()
	done := make(chan error, 1)
	go func() {
		done <- forward(context.Background(), transport, Tunnel{Endpoint: "https://localhost:" + strings.Split(packet.LocalAddr().String(), ":")[1], Access: "team/access", Token: proofToken}, local)
	}()
	_, _ = client.Write([]byte("hello"))
	_ = client.(*net.TCPConn).CloseWrite()
	data, err := io.ReadAll(client)
	if err != nil || string(data) != "reply:hello" {
		t.Fatalf("reply %q: %v", data, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	deniedClient, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = deniedClient.Close() }()
	deniedLocal, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = deniedLocal.Close() }()
	denied := make(chan error, 1)
	go func() {
		denied <- forward(context.Background(), transport, Tunnel{Endpoint: "https://localhost:" + strings.Split(packet.LocalAddr().String(), ":")[1], Access: "team/access", Token: deniedToken}, deniedLocal)
	}()
	select {
	case err := <-denied:
		if err == nil || !strings.Contains(err.Error(), "401") {
			t.Fatalf("expected 401, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("denied stream did not close")
	}
}

func TestRequestRejectsUntrustedEndpoint(t *testing.T) {
	for _, endpoint := range []string{"http://relay", "https://relay/path", "https://user@relay", "https://relay?target=db"} {
		if _, err := (Tunnel{Endpoint: endpoint, Access: "team/access", Token: proofToken}).request(context.Background(), nil); err == nil {
			t.Errorf("accepted %q", endpoint)
		}
	}
}

func TestRequestRejectsMalformedTokensWithoutEchoingThem(t *testing.T) {
	for name, token := range map[string]string{
		"empty":         "",
		"short":         "c2hvcnQ",
		"padded":        base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)),
		"newline":       proofToken + "\n",
		"not-base64url": strings.Repeat("!", 43),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := (Tunnel{Endpoint: "https://relay.example:8443", Access: "team/access", Token: token}).request(context.Background(), nil)
			if err == nil {
				t.Fatal("expected malformed token to be rejected")
			}
			if token != "" && strings.Contains(err.Error(), token) {
				t.Fatalf("error leaks token: %v", err)
			}
		})
	}
}
