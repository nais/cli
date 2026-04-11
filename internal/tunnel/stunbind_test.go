package tunnel

import (
	"net"
	"testing"
)

func TestSTUNBindOpen(t *testing.T) {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	expectedPort := uint16(udpConn.LocalAddr().(*net.UDPAddr).Port)

	bind := NewSTUNBind(udpConn)
	recvFuncs, port, err := bind.Open(0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if len(recvFuncs) == 0 {
		t.Error("expected non-empty receive functions slice")
	}
	if port != expectedPort {
		t.Errorf("expected port %d, got %d", expectedPort, port)
	}

	bind.Close()
}

func TestSTUNBindClose(t *testing.T) {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}

	bind := NewSTUNBind(udpConn)
	_, _, err = bind.Open(0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := bind.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}

	if err := bind.Close(); err != nil {
		t.Errorf("second Close returned unexpected error: %v", err)
	}
}

func TestSTUNBindParseEndpoint(t *testing.T) {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer udpConn.Close()

	bind := NewSTUNBind(udpConn)

	input := "1.2.3.4:5678"
	ep, err := bind.ParseEndpoint(input)
	if err != nil {
		t.Fatalf("ParseEndpoint(%q): %v", input, err)
	}
	if ep == nil {
		t.Fatal("ParseEndpoint returned nil endpoint")
	}
	got := ep.DstToString()
	if got != input {
		t.Errorf("DstToString() = %q, want %q", got, input)
	}
}
