package tunnel

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	// TunnelIPClient is the WireGuard tunnel IP for the CLI side.
	TunnelIPClient = "10.0.0.1/30"
	// TunnelIPGateway is the WireGuard tunnel IP for the gateway side.
	TunnelIPGateway = "10.0.0.2/30"
	// PersistentKeepalive in seconds — keeps NAT mappings alive (< Cloud NAT 30s timeout).
	PersistentKeepalive = 20
)

// WireGuardTunnel wraps a userspace WireGuard device with an associated netstack.
type WireGuardTunnel struct {
	dev *device.Device
	net *netstack.Net
}

// SetupWireGuard creates a userspace WireGuard device for the CLI side of the tunnel.
// privateKey: CLI's WireGuard private key
// gatewayPublicKey: gateway's WireGuard public key
// gatewayEndpoint: gateway's UDP endpoint (ip:port) exposed by the forwarder
func SetupWireGuard(privateKey wgtypes.Key, gatewayPublicKey wgtypes.Key, gatewayEndpoint string) (*WireGuardTunnel, error) {
	return setupWireGuard(privateKey, gatewayPublicKey, gatewayEndpoint, conn.NewDefaultBind(), "listen_port=0\n")
}

func setupWireGuard(privateKey wgtypes.Key, gatewayPublicKey wgtypes.Key, gatewayEndpoint string, bind conn.Bind, listenPortConfig string) (*WireGuardTunnel, error) {
	prefix, err := netip.ParsePrefix(TunnelIPClient)
	if err != nil {
		return nil, fmt.Errorf("parse tunnel IP: %w", err)
	}

	tun, net, err := netstack.CreateNetTUN(
		[]netip.Addr{prefix.Addr()},
		[]netip.Addr{}, // no DNS
		1420,           // MTU
	)
	if err != nil {
		return nil, fmt.Errorf("create netstack TUN: %w", err)
	}
	if bind == nil {
		return nil, fmt.Errorf("wireguard bind is nil")
	}

	logger := device.NewLogger(device.LogLevelError, "[wireguard-cli] ")
	dev := device.NewDevice(tun, bind, logger)

	cfg := fmt.Sprintf(`private_key=%s
%spublic_key=%s
persistent_keepalive_interval=%d
allowed_ip=0.0.0.0/0
endpoint=%s
`, encodeKeyHex(privateKey), listenPortConfig, encodeKeyHex(gatewayPublicKey), PersistentKeepalive, gatewayEndpoint)

	if err := dev.IpcSet(cfg); err != nil {
		dev.Close()
		return nil, fmt.Errorf("configure wireguard: %w", err)
	}

	if err := dev.Up(); err != nil {
		dev.Close()
		return nil, fmt.Errorf("bring up wireguard: %w", err)
	}

	return &WireGuardTunnel{dev: dev, net: net}, nil
}

// Net returns the netstack network for creating TCP connections through the tunnel.
func (t *WireGuardTunnel) Net() *netstack.Net {
	return t.net
}

// DialTCP creates a TCP connection through the WireGuard tunnel.
func (t *WireGuardTunnel) DialTCP(addr string) (net.Conn, error) {
	return t.net.Dial("tcp", addr)
}

// Close shuts down the WireGuard device cleanly.
func (t *WireGuardTunnel) Close() {
	if t != nil && t.dev != nil {
		t.dev.Close()
	}
}

func encodeKeyHex(key wgtypes.Key) string {
	return hex.EncodeToString(key[:])
}
