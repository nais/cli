package tunnel

import (
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"

	"golang.zx2c4.com/wireguard/conn"
)

type STUNBind struct {
	conn      *net.UDPConn
	closeOnce sync.Once
}

type StdNetEndpoint net.UDPAddr

func NewSTUNBind(udpConn *net.UDPConn) *STUNBind {
	return &STUNBind{conn: udpConn}
}

func (b *STUNBind) Open(port uint16) ([]conn.ReceiveFunc, uint16, error) {
	_ = port
	if b == nil || b.conn == nil {
		return nil, 0, fmt.Errorf("nil UDP connection")
	}

	localAddr, ok := b.conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddr == nil {
		return nil, 0, fmt.Errorf("local address is not UDP")
	}

	recv := func(packets [][]byte, sizes []int, eps []conn.Endpoint) (int, error) {
		if len(packets) == 0 || len(sizes) == 0 || len(eps) == 0 {
			return 0, io.ErrShortBuffer
		}

		n, addr, err := b.conn.ReadFromUDP(packets[0])
		if err != nil {
			return 0, err
		}

		sizes[0] = n
		eps[0] = newStdNetEndpoint(addr)
		return 1, nil
	}

	return []conn.ReceiveFunc{recv}, uint16(localAddr.Port), nil
}

func (b *STUNBind) Close() error {
	if b == nil || b.conn == nil {
		return nil
	}

	var err error
	b.closeOnce.Do(func() {
		err = b.conn.Close()
	})
	return err
}

func (b *STUNBind) SetMark(mark uint32) error {
	_ = mark
	return nil
}

func (b *STUNBind) Send(bufs [][]byte, ep conn.Endpoint) error {
	if b == nil || b.conn == nil {
		return fmt.Errorf("nil UDP connection")
	}

	addr, err := udpAddrFromEndpoint(ep)
	if err != nil {
		return err
	}

	for _, buf := range bufs {
		if _, err := b.conn.WriteToUDP(buf, addr); err != nil {
			return err
		}
	}

	return nil
}

func (b *STUNBind) ParseEndpoint(s string) (conn.Endpoint, error) {
	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		return nil, err
	}
	return newStdNetEndpoint(addr), nil
}

func (b *STUNBind) BatchSize() int {
	return 1
}

func (e *StdNetEndpoint) ClearSrc() {}

func (e *StdNetEndpoint) SrcToString() string {
	return ""
}

func (e *StdNetEndpoint) DstToString() string {
	addr := (*net.UDPAddr)(e)
	if addr == nil {
		return ""
	}
	return addr.String()
}

func (e *StdNetEndpoint) DstToBytes() []byte {
	addr := (*net.UDPAddr)(e)
	if addr == nil {
		return nil
	}
	b, err := addr.AddrPort().MarshalBinary()
	if err != nil {
		return nil
	}
	return b
}

func (e *StdNetEndpoint) DstIP() netip.Addr {
	return udpAddrIP((*net.UDPAddr)(e))
}

func (e *StdNetEndpoint) SrcIP() netip.Addr {
	return netip.Addr{}
}

func newStdNetEndpoint(addr *net.UDPAddr) *StdNetEndpoint {
	if addr == nil {
		return nil
	}

	clone := *addr
	if addr.IP != nil {
		clone.IP = append(net.IP(nil), addr.IP...)
	}

	return (*StdNetEndpoint)(&clone)
}

func udpAddrFromEndpoint(ep conn.Endpoint) (*net.UDPAddr, error) {
	switch e := ep.(type) {
	case *StdNetEndpoint:
		if e == nil {
			return nil, fmt.Errorf("nil endpoint")
		}
		addr := (*net.UDPAddr)(e)
		clone := *addr
		if addr.IP != nil {
			clone.IP = append(net.IP(nil), addr.IP...)
		}
		return &clone, nil
	case interface{ DstToString() string }:
		addr, err := net.ResolveUDPAddr("udp", e.DstToString())
		if err != nil {
			return nil, fmt.Errorf("resolve endpoint %q: %w", e.DstToString(), err)
		}
		return addr, nil
	default:
		return nil, fmt.Errorf("unsupported endpoint type %T", ep)
	}
}

func udpAddrIP(addr *net.UDPAddr) netip.Addr {
	if addr == nil {
		return netip.Addr{}
	}

	ip, ok := netip.AddrFromSlice(addr.IP)
	if !ok {
		return netip.Addr{}
	}

	return ip.Unmap()
}
