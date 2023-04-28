package http

import (
	"net"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/proto/tcp"
	"github.com/siketyan/nakoud/pkg/proxy"
)

type Proxy struct {
	addr *net.TCPAddr
}

func NewProxy(address string) (*Proxy, error) {
	addr, err := net.ResolveTCPAddr(tcp.NetworkTcp, address)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		addr: addr,
	}, nil
}

func (p *Proxy) AsProxy() proxy.Proxy {
	return p
}

func (p *Proxy) Run() error {
	listener, err := net.ListenTCP(tcp.NetworkTcp, p.addr)
	if err != nil {
		return err
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
