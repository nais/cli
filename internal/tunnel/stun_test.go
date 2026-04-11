package tunnel

import (
	"net"
	"regexp"
	"strings"
	"testing"

	"github.com/pion/stun/v3"
)

func TestDiscoverSTUNEndpoint(t *testing.T) {
	endpoint, conn, err := DiscoverSTUNEndpoint(0)
	if err != nil {
		t.Fatalf("discover STUN endpoint: %v", err)
	}
	defer conn.Close()

	pattern := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+:\d+$`)
	if !pattern.MatchString(endpoint) {
		t.Errorf("endpoint %q does not match x.x.x.x:port pattern", endpoint)
	}
	t.Logf("Discovered STUN endpoint: %s", endpoint)
}

func startMockSTUNServer(t *testing.T, mappedIP string, mappedPort int) string {
	t.Helper()
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { udpConn.Close() })

	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			req := &stun.Message{Raw: make([]byte, n)}
			copy(req.Raw, buf[:n])
			if err := req.Decode(); err != nil {
				continue
			}
			resp, err := stun.Build(
				stun.NewTransactionIDSetter(req.TransactionID),
				stun.NewType(stun.MethodBinding, stun.ClassSuccessResponse),
				&stun.XORMappedAddress{
					IP:   net.ParseIP(mappedIP),
					Port: mappedPort,
				},
				stun.Fingerprint,
			)
			if err != nil {
				continue
			}
			udpConn.WriteTo(resp.Raw, addr)
		}
	}()

	return udpConn.LocalAddr().String()
}

func TestSymmetricNATDetection(t *testing.T) {
	server1 := startMockSTUNServer(t, "1.2.3.4", 10000)
	server2 := startMockSTUNServer(t, "1.2.3.4", 20000)

	localConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer localConn.Close()

	_, err = discoverFromServers(localConn, []string{server1, server2})
	if err == nil {
		t.Fatal("expected symmetric NAT error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "symmetric nat") {
		t.Errorf("expected error to mention symmetric NAT, got: %v", err)
	}
}

func TestAllSTUNServersFailed(t *testing.T) {
	deadConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	deadServer := deadConn.LocalAddr().String()
	deadConn.Close()

	localConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer localConn.Close()

	_, err = discoverFromServers(localConn, []string{deadServer})
	if err == nil {
		t.Fatal("expected error when all STUN servers fail, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "all stun servers failed") {
		t.Errorf("expected 'all STUN servers failed' in error, got: %v", err)
	}
}

func TestSTUNEndpointFormat(t *testing.T) {
	server := startMockSTUNServer(t, "5.6.7.8", 12345)

	localConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer localConn.Close()

	endpoint, err := discoverFromServer(localConn, server)
	if err != nil {
		t.Fatalf("discoverFromServer: %v", err)
	}

	pattern := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+:\d+$`)
	if !pattern.MatchString(endpoint) {
		t.Errorf("endpoint %q does not match ip:port format", endpoint)
	}
	if endpoint != "5.6.7.8:12345" {
		t.Errorf("expected endpoint 5.6.7.8:12345, got %q", endpoint)
	}
}
