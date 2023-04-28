package tcp

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/upstream"
)

type Client struct {
	address *net.TCPAddr
}

func NewClient(address *net.TCPAddr) *Client {
	return &Client{
		address: address,
	}
}

func (c *Client) Connect() (upstream.Connection, error) {
	log.Info().Any("Address", c.address).Msg("Connecting to a TCP upstream")

	conn, err := net.DialTCP("tcp", nil, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to start a TCP connection: %w", err)
	}

	return &Connection{
		conn: conn,
	}, nil
}
