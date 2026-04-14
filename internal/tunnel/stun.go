package tunnel

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v3"
)

// defaultSTUNServers is the ordered list of STUN servers.
// Cloudflare is primary (anycast, Oslo/Stockholm PoP), Google servers are fallback.
var defaultSTUNServers = []string{
	"stun.cloudflare.com:3478",
	"stun.l.google.com:19302",
	"stun1.l.google.com:19302",
	"stun2.l.google.com:19302",
}

// DiscoverSTUNEndpoint performs STUN discovery to find the external UDP endpoint.
// Returns the discovered endpoint (ip:port string) and the UDP connection that was used.
// The caller must keep the connection open for hole-punching — close it only after the tunnel is established.
//
// Returns an error if:
// - All STUN servers fail (network issue)
// - Symmetric NAT detected (different endpoints returned by different servers)
func DiscoverSTUNEndpoint(listenPort int) (endpoint string, conn *net.UDPConn, err error) {
	conn, err = net.ListenUDP("udp4", &net.UDPAddr{Port: listenPort})
	if err != nil {
		return "", nil, fmt.Errorf("listen UDP for STUN: %w", err)
	}

	endpoint, err = discoverFromServers(conn, defaultSTUNServers)
	if err != nil {
		conn.Close() // #nosec G104 -- best-effort cleanup on error path
		return "", nil, err
	}

	return endpoint, conn, nil
}

func discoverFromServers(conn *net.UDPConn, servers []string) (string, error) {
	var lastErr error
	var firstEndpoint string

	for i, server := range servers {
		endpoint, err := discoverFromServer(conn, server)
		if err != nil {
			lastErr = fmt.Errorf("STUN server %s: %w", server, err)
			continue
		}

		if i == 0 {
			firstEndpoint = endpoint
			continue
		}

		// Check for symmetric NAT: if two different servers report different ports,
		// we have symmetric NAT and cannot hole-punch.
		if endpoint != firstEndpoint {
			return "", fmt.Errorf(
				"your network uses symmetric NAT which prevents direct tunnel connections. "+
					"Try from a different network or disable VPN. "+
					"(server 1 reported %s, server 2 reported %s)",
				firstEndpoint, endpoint,
			)
		}
		return firstEndpoint, nil
	}

	// If we only tried one server successfully, return that
	if firstEndpoint != "" {
		return firstEndpoint, nil
	}

	return "", fmt.Errorf("all STUN servers failed (check network connectivity): %w", lastErr)
}

func discoverFromServer(conn *net.UDPConn, server string) (string, error) {
	serverAddr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return "", fmt.Errorf("set deadline: %w", err)
	}
	defer conn.SetDeadline(time.Time{})

	if _, err := conn.WriteTo(message.Raw, serverAddr); err != nil {
		return "", fmt.Errorf("send STUN request: %w", err)
	}

	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", fmt.Errorf("read STUN response: %w", err)
	}

	resp := &stun.Message{Raw: buf[:n]}
	if err := resp.Decode(); err != nil {
		return "", fmt.Errorf("decode STUN response: %w", err)
	}

	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(resp); err != nil {
		var mappedAddr stun.MappedAddress
		if err2 := mappedAddr.GetFrom(resp); err2 != nil {
			return "", fmt.Errorf("get mapped address: %w", err)
		}
		return fmt.Sprintf("%s:%d", mappedAddr.IP.String(), mappedAddr.Port), nil
	}

	return fmt.Sprintf("%s:%d", xorAddr.IP.String(), xorAddr.Port), nil
}
