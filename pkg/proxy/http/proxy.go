package http

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/proto/tcp"
)

type Proxy struct {
	addr *net.TCPAddr
}

func NewProxy(address string) (*Proxy, error) {
	addr, err := net.ResolveTCPAddr(tcp.NetworkTCP, address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the TCP address: %w", err)
	}

	return &Proxy{
		addr: addr,
	}, nil
}

func (p *Proxy) Run() error {
	listener, err := net.ListenTCP(tcp.NetworkTCP, p.addr)
	if err != nil {
		return fmt.Errorf("failed to start listening on a TCP socket: %w", err)
	}

	defer func() {
		_ = listener.Close()
	}()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Error().Err(err).Msg("Could not accept the connection")

			continue
		}

		go func() {
			if err := (&Connection{conn: conn}).Handle(); err != nil {
				log.Error().Err(err).Msg("An error occurred while processing a request")
			}
		}()
	}
}
