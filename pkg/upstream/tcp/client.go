package tcp

import (
	"net"

	"github.com/rs/zerolog/log"
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) Connect() (*Connection, error) {
	log.Info().Str("Address", c.addr).Msg("Connecting to a TCP upstream")

	addr, err := net.ResolveTCPAddr("tcp", c.addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn: conn,
	}, nil
}
