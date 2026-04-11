package tunnel

import (
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestSetupWireGuard(t *testing.T) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	peerKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate peer key: %v", err)
	}

	tun, err := SetupWireGuard(privKey, peerKey.PublicKey(), "127.0.0.1:51820")
	if err != nil {
		t.Fatalf("setup wireguard: %v", err)
	}
	defer tun.Close()

	if tun.Net() == nil {
		t.Error("expected non-nil net")
	}
}
