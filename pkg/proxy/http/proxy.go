package http

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/discovery"
	"github.com/siketyan/nakoud/pkg/proto/tcp"
)

type Proxy struct {
	addr       *net.TCPAddr
	discoverer discovery.Discoverer
}

func NewProxy(address string, discoverer discovery.Discoverer) (*Proxy, error) {
	addr, err := net.ResolveTCPAddr(tcp.NetworkTCP, address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the TCP address: %w", err)
	}

	return &Proxy{
		addr:       addr,
		discoverer: discoverer,
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
		connection, err := listener.AcceptTCP()
		if err != nil {
			log.Error().Err(err).Msg("Could not accept the connection")

			continue
		}

		go func() {
			ctx := context.Background()
			if err := NewConnection(connection, p.discoverer).Handle(ctx); err != nil {
				log.Error().Err(err).Msg("An error occurred while processing a request")
			}
		}()
	}
}
