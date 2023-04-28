package ip

import (
	"fmt"
	"net"
	"net/netip"
)

type Address struct {
	inner   netip.Addr
	network *net.IPNet
}

func ParseAddress(address string, network *net.IPNet) (*Address, error) {
	ipAddress, err := netip.ParseAddr(address)
	if err != nil {
		return nil, fmt.Errorf("malformed IP address: %w", err)
	}

	return &Address{
		inner:   ipAddress,
		network: network,
	}, nil
}

func (a *Address) AsTCP(port uint16) *net.TCPAddr {
	return net.TCPAddrFromAddrPort(netip.AddrPortFrom(a.inner, port))
}
